package upload

import (
	"bytes"
	"code-sync-client/misc"
	"github.com/goinbox/gomisc"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/goinbox/crypto"
	"github.com/goinbox/gohttp/httpclient"

	"code-sync-client/conf"
	"code-sync-client/resource"
	"code-sync-client/svc"
)

const (
	MultiPartFormNameMd5     = "md5"
	MultiPartFormNameFile    = "formfile"
	MultiPartFormNamePerm    = "perm"
	MultiPartFormNameVersion = "version"
)

var fileSignQueryNames = append([]string{"prj", "user", "host", MultiPartFormNameFile, MultiPartFormNameMd5, MultiPartFormNamePerm, MultiPartFormNameVersion}, misc.ApiSignQueryNames...)

type UploadSvc struct {
	*svc.BaseSvc
}

func NewUploadSvc(traceId []byte) *UploadSvc {
	return &UploadSvc{
		BaseSvc: &svc.BaseSvc{
			TraceId: traceId,
		},
	}
}

func (us *UploadSvc) UploadFile(cpc *conf.CodePrjConf, rpath string, version int) error {
	req, err := us.makeUploadFileRequest(cpc, rpath, version)
	if err != nil {
		us.ErrorLog([]byte("UploadFile"), []byte("makeUploadFileRequest"))
		return err
	}

	client := httpclient.NewClient(httpclient.NewConfig(), resource.AccessLogger)
	resp, err := client.Do(req, 1)
	if err != nil {
		us.ErrorLog([]byte("UploadFile"), []byte("RequestServer"))
		return err
	}

	us.InfoLog([]byte("UploadFile.Response"), resp.Contents)

	return nil
}

func (us *UploadSvc) makeUploadFileRequest(cpc *conf.CodePrjConf, rpath string, version int) (*httpclient.Request, error) {
	bodyBuffer := bytes.NewBuffer([]byte{})
	writer := multipart.NewWriter(bodyBuffer)

	extHeaders := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	apath := cpc.PrjHome + "/" + rpath

	originPartData := make(map[string]string)
	err := us.makeMultipartPerm(apath, writer, originPartData)
	if err != nil {
		us.ErrorLog([]byte("makeMultipartPerm"), []byte(err.Error()))
		return nil, err
	}

	err = us.makeMultipartVersion(version, writer, originPartData)
	if err != nil {
		us.ErrorLog([]byte("makeUploadFileRequest"), []byte("WriteVersionField"))
		return nil, err
	}

	err = us.makeMultipartFile(apath, rpath, writer, originPartData)
	if err != nil {
		us.ErrorLog([]byte("makeUploadFileRequest"), []byte("makeMultipartFile"))
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		us.ErrorLog([]byte("makeUploadFileRequest"), []byte("CloseWriter"))
		return nil, err
	}

	return httpclient.NewRequest(http.MethodPost, us.makeUploadFileRequestUrl(cpc, originPartData), bodyBuffer.Bytes(), "", extHeaders)
}

func (us *UploadSvc) makeMultipartPerm(path string, writer *multipart.Writer, originPartData map[string]string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	permStr := strconv.Itoa(int(fi.Mode().Perm()))
	err = writer.WriteField(MultiPartFormNamePerm, permStr)
	if err != nil {
		return err
	}

	originPartData[MultiPartFormNamePerm] = permStr

	return nil
}

func (us *UploadSvc) makeMultipartVersion(version int, writer *multipart.Writer, originPartData map[string]string) error {
	versionStr := strconv.Itoa(version)
	err := writer.WriteField(MultiPartFormNameVersion, versionStr)
	if err != nil {
		return err
	}

	originPartData[MultiPartFormNameVersion] = versionStr

	return nil
}

func (us *UploadSvc) makeMultipartFile(apath, rpath string, writer *multipart.Writer, originPartData map[string]string) error {
	contents, err := ioutil.ReadFile(apath)
	if err != nil {
		return err
	}

	md5Str := crypto.Md5String(contents)
	err = writer.WriteField(MultiPartFormNameMd5, md5Str)
	if err != nil {
		return err
	}

	originPartData[MultiPartFormNameMd5] = md5Str

	part, err := writer.CreateFormFile(MultiPartFormNameFile, rpath)
	if err != nil {
		return err
	}

	originPartData[MultiPartFormNameFile] = rpath

	_, err = part.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

func (us *UploadSvc) makeUploadFileRequestUrl(cpc *conf.CodePrjConf, originPartData map[string]string) string {
	vs := url.Values{}
	vs.Set("prj", cpc.PrjName)
	vs.Set("user", conf.BaseConf.Username)
	vs.Set("host", conf.BaseConf.Hostname)

	ru := "http://" + cpc.CodeSyncServer.Host + ":" + cpc.CodeSyncServer.Port
	ru += "/upload/file?"
	ru += vs.Encode()

	for _, key := range fileSignQueryNames {
		v, ok := originPartData[key]
		if ok {
			vs.Set(key, v)
		}
	}
	ru += "&" + us.makeSignParams(vs, fileSignQueryNames, cpc.Token)

	return ru
}

func (us *UploadSvc) makeSignParams(queryValues url.Values, signQueryNames []string, token string) string {
	now := time.Now()
	t := strconv.FormatInt(now.Unix(), 10)
	nonce := strconv.FormatInt(gomisc.RandByTime(&now), 10)

	queryValues.Set("t", t)
	queryValues.Set("nonce", nonce)

	r := "t=" + t
	r += "&nonce=" + nonce
	r += "&sign=" + misc.CalApiSign(queryValues, signQueryNames, token)

	return r
}
