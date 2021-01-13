package sync

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/goinbox/crypto"
	"github.com/goinbox/gohttp/httpclient"
	"github.com/goinbox/gomisc"

	"code-sync-client/conf"
	"code-sync-client/misc"
	"code-sync-client/resource"
	"code-sync-client/svc"
)

const (
	MultiPartFormNameMd5  = "md5"
	MultiPartFormNameFile = "formfile"
	MultiPartFormNamePerm = "perm"

	RequestRetry = 3
)

var uploadSignQueryNames = append([]string{"prj", "user", "host", MultiPartFormNameFile, MultiPartFormNameMd5, MultiPartFormNamePerm}, misc.ApiSignQueryNames...)
var deleteSignQueryNames = append([]string{"prj", "user", "host", "rpath"}, misc.ApiSignQueryNames...)

type SyncSvc struct {
	*svc.BaseSvc

	client *httpclient.Client
}

func NewSyncSvc(traceId []byte) *SyncSvc {
	return &SyncSvc{
		BaseSvc: &svc.BaseSvc{
			TraceId: traceId,
		},
		client: httpclient.NewClient(httpclient.NewConfig(), resource.AccessLogger),
	}
}

func (ss *SyncSvc) UploadFile(cpc *conf.CodePrjConf, rpath string) error {
	apath := cpc.PrjHome + "/" + rpath
	var rpathList []string
	if gomisc.DirExist(apath) {
		files, _ := gomisc.ListFilesInDir(apath)
		for _, apath := range files {
			rpathList = append(rpathList, misc.RelativePath(cpc.PrjHome, apath))
		}
	} else {
		rpathList = append(rpathList, rpath)
	}

	for _, fpath := range rpathList {
		if misc.PathInExcludeList(fpath, cpc.ExcludeList) {
			ss.InfoLog([]byte("UploadFile"), []byte("exclude "+fpath))
			continue
		}

		requestList, err := ss.makeUploadFileRequestList(cpc, fpath)
		if err != nil {
			ss.ErrorLog([]byte("UploadFile"), []byte("makeUploadFileRequest"))
			continue
		}

		ss.InfoLog([]byte("UploadFile"), []byte("upload "+fpath))
		ss.request(requestList)
	}

	return nil
}

func (ss *SyncSvc) DeleteFile(cpc *conf.CodePrjConf, rpath string) error {
	if misc.PathInExcludeList(rpath, cpc.ExcludeList) {
		ss.InfoLog([]byte("DeleteFile"), []byte("exclude "+rpath))
		return nil
	}

	requestList, err := ss.makeDeleteFileRequestList(cpc, rpath)
	if err != nil {
		ss.ErrorLog([]byte("DeleteFile"), []byte("makeDeleteFileRequest"))
		return err
	}

	ss.InfoLog([]byte("DeleteFile"), []byte("delete "+rpath))
	ss.request(requestList)

	return nil
}

func (ss *SyncSvc) makeUploadFileRequestList(cpc *conf.CodePrjConf, rpath string) ([]*httpclient.Request, error) {
	bodyBuffer := bytes.NewBuffer([]byte{})
	writer := multipart.NewWriter(bodyBuffer)

	extHeaders := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	apath := cpc.PrjHome + "/" + rpath

	originPartData := make(map[string]string)
	err := ss.makeMultipartPerm(apath, writer, originPartData)
	if err != nil {
		ss.ErrorLog([]byte("makeMultipartPerm"), []byte(err.Error()))
		return nil, err
	}

	err = ss.makeMultipartFile(apath, rpath, writer, originPartData)
	if err != nil {
		ss.ErrorLog([]byte("makeUploadFileRequest"), []byte("makeMultipartFile "+rpath+" "+err.Error()))
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		ss.ErrorLog([]byte("makeUploadFileRequest"), []byte("CloseWriter"))
		return nil, err
	}

	vs := ss.queryValues(cpc, nil)
	var requestList []*httpclient.Request
	for _, serverConf := range cpc.CodeSyncServerList {
		ru := ss.makeRequestUrl("file/upload", vs, serverConf, uploadSignQueryNames, originPartData)
		request, err := httpclient.NewRequest(http.MethodPost, ru, bodyBuffer.Bytes(), "", extHeaders)
		if err != nil {
			return nil, err
		}
		requestList = append(requestList, request)
	}

	return requestList, nil
}

func (ss *SyncSvc) makeMultipartPerm(path string, writer *multipart.Writer, originPartData map[string]string) error {
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

func (ss *SyncSvc) makeMultipartFile(apath, rpath string, writer *multipart.Writer, originPartData map[string]string) error {
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

func (ss *SyncSvc) makeDeleteFileRequestList(cpc *conf.CodePrjConf, rpath string) ([]*httpclient.Request, error) {
	signValues := map[string]string{
		"rpath": rpath,
	}
	vs := ss.queryValues(cpc, signValues)

	var requestList []*httpclient.Request
	for _, serverConf := range cpc.CodeSyncServerList {
		ru := ss.makeRequestUrl("file/delete", vs, serverConf, deleteSignQueryNames, signValues)
		request, err := httpclient.NewRequest(http.MethodGet, ru, nil, "", nil)
		if err != nil {
			return nil, err
		}
		requestList = append(requestList, request)
	}

	return requestList, nil
}

func (ss *SyncSvc) queryValues(cpc *conf.CodePrjConf, extValues map[string]string) url.Values {
	vs := url.Values{}

	vs.Set("prj", cpc.PrjName)
	vs.Set("user", conf.CommonConf.Username)
	vs.Set("host", conf.CommonConf.Hostname)

	if extValues != nil {
		for k, v := range extValues {
			vs.Set(k, v)
		}
	}

	return vs
}

func (ss *SyncSvc) makeRequestUrl(controllerAction string, vs url.Values, serverConf *conf.CodeSyncServerConf, signQueryNames []string, signQueryValues map[string]string) string {
	ru := "http://" + serverConf.Host + ":" + serverConf.Port
	ru += serverConf.Path + controllerAction + "?"
	ru += vs.Encode()

	for _, key := range signQueryNames {
		v, ok := signQueryValues[key]
		if ok {
			vs.Set(key, v)
		}
	}
	ru += "&" + ss.makeSignParams(vs, signQueryNames, serverConf.Token)

	return ru
}

func (ss *SyncSvc) makeSignParams(queryValues url.Values, signQueryNames []string, token string) string {
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

func (ss *SyncSvc) request(requestList []*httpclient.Request) {
	for _, req := range requestList {
		resp, err := ss.client.Do(req, RequestRetry)
		if err != nil {
			ss.ErrorLog([]byte("request"), []byte("RequestServerError: "+err.Error()))
		} else {
			ss.InfoLog([]byte("request.Response"), resp.Contents)
		}
	}
}
