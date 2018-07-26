package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qiniu/api.v6/auth/digest"
	"github.com/qiniu/log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"ufop"
	"ufop/utils"
)

/*
ffmpeg
/format/<video format>
/wmImage/<encoded image url>
/wmOffsetX/<offsetX>
/wmOffsetY/<offsetY>
*/
type FFmpeg struct {
	mac *digest.Mac
}

type FFmpegConfig struct {
	//ak & sk
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

func (this *FFmpeg) Name() string {
	return "ffmpeg"
}

func (this *FFmpeg) InitConfig(jobConf string) (err error) {
	confFp, openErr := os.Open(jobConf)
	if openErr != nil {
		err = fmt.Errorf("Open ffmpeg config failed, %s", openErr.Error())
		return
	}

	config := FFmpegConfig{}
	decoder := json.NewDecoder(confFp)
	decodeErr := decoder.Decode(&config)
	if decodeErr != nil {
		err = fmt.Errorf("Parse ffmpeg config failed, %s", decodeErr.Error())
		return
	}

	this.mac = &digest.Mac{config.AccessKey, []byte(config.SecretKey)}

	return
}

func (this *FFmpeg) parse(cmd string) (format, wmImage, wmOffsetX, wmOffsetY string, err error) {
	pattern := "^ffmpeg/format/[0-9a-zA-Z-_=]+/wmImage/[0-9a-zA-Z-_=]+/wmOffsetX/[0-9]+/wmOffsetY/[0-9]+"
	matched, _ := regexp.MatchString(pattern, cmd)
	if !matched {
		err = errors.New("invalid ffmpeg command format")
		return
	}

	var decodeErr error
	//get image
	wmImage, decodeErr = utils.GetParamDecoded(cmd, "wmImage/[0-9a-zA-Z-_=]+", "wmImage")
	if decodeErr != nil {
		err = errors.New("invalid ffmpeg paramter 'wmImage'")
		return
	}

	format = utils.GetParam(cmd, "format/[0-9a-zA-Z-_=]+", "format")
	wmOffsetX = utils.GetParam(cmd, "wmOffsetX/[0-9]+", "wmOffsetX")
	wmOffsetY = utils.GetParam(cmd, "wmOffsetY/[0-9]+", "wmOffsetY")

	return
}

func (this *FFmpeg) Do(req ufop.UfopRequest, ufopBody io.ReadCloser) (result interface{}, resultType int, contentType string, err error) {
	reqId := req.ReqId
	//parse command
	format, wmImageUrl, wmOffsetX, wmOffsetY, pErr := this.parse(req.Cmd)
	if pErr != nil {
		err = pErr
		return
	}

	//download video
	videoUrl := req.Url
	resResp, respErr := http.Get(videoUrl)
	if respErr != nil || resResp.StatusCode != 200 {
		if respErr != nil {
			err = fmt.Errorf("retrieve resource video failed, %s", respErr.Error())
		} else {
			err = fmt.Errorf("retrieve resource video failed, %s", resResp.Status)
			if resResp.Body != nil {
				resResp.Body.Close()
			}
		}
		return
	}
	defer resResp.Body.Close()

	videoPath := fmt.Sprintf("video_%s", reqId)
	f, fErr := os.Create(videoPath)
	if fErr != nil {
		err = fmt.Errorf("create save video file failed, %s", fErr.Error())
		return
	}
	_, cpErr := io.Copy(f, resResp.Body)
	if cpErr != nil {
		err = fmt.Errorf("save local video file failed, %s", cpErr.Error())
		return
	}
	defer f.Close()

	//download wmImage
	resResp, respErr = http.Get(wmImageUrl)
	if respErr != nil || resResp.StatusCode != 200 {
		if respErr != nil {
			err = fmt.Errorf("retrieve resource image failed, %s", respErr.Error())
		} else {
			err = fmt.Errorf("retrieve resource image failed, %s", resResp.Status)
			if resResp.Body != nil {
				resResp.Body.Close()
			}
		}
		return
	}
	defer resResp.Body.Close()

	wmImagePath := fmt.Sprintf("wmImage_%s", reqId)
	f, fErr = os.Create(wmImagePath)
	if fErr != nil {
		err = fmt.Errorf("create save image file failed, %s", fErr.Error())
		return
	}
	_, cpErr = io.Copy(f, resResp.Body)
	if cpErr != nil {
		err = fmt.Errorf("save local image file failed, %s", cpErr.Error())
		return
	}
	defer f.Close()

	//handle video file
	//ffmpeg -i input.mp4 -i wm.png -filter_complex "overlay=20:40" output.mp4
	// fmt.Printf("wmOffsetX = %s\n", wmOffsetX)
	// fmt.Printf("wmOffsetY = %s\n", wmOffsetY)
	// absPath, _ := os.Getwd()
	// fmt.Println("current path:", absPath)

	// outFile, createErr := os.Create(fmt.Sprintf("ouput_%s.%s", reqId, format))
	// if createErr != nil {
	// 	err = fmt.Errorf("creat output file failed,%s", createErr.Error())
	// 	return
	// }

	outFile := fmt.Sprintf("ouput_%s.%s", reqId, format)
	cmdStr := fmt.Sprintf("ffmpeg -i %s -i %s -filter_complex 'overlay=%s:%s' %s", videoPath, wmImagePath, wmOffsetX, wmOffsetY, outFile)
	fmt.Println("ffmpeg", "-i", cmdStr)
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		err = fmt.Errorf("exec ffmpeg command failed, %s", cmdErr.Error())
		return
	}
	defer os.Remove(videoPath)
	defer os.Remove(wmImagePath)

	//write result
	result = outFile
	resultType = ufop.RESULT_TYPE_OCTET_FILE
	contentType = ufop.CONTENT_TYPE_OCTET

	log.Infof("[%s] ffmpeg success!", reqId)
	return
}
