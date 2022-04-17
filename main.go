package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

type infoStruct struct {
	ArticleName string `json:"articlename"`
	Author      string `json:"author"`
	Intro       string `json:"intro"`
}

type listStruct struct {
	ChapterId   string `json:"chapterid"`
	ChapterName string `json:"chaptername"`
}

type volumeStruct struct {
	Url       string       `json:"url"`
	Info      infoStruct   `json:"info"`
	List      []listStruct `json:"list"`
	ArticleId int          `json:"article_id"`
}

type Urls struct {
	url      string
	fileName string
}

func main() {
	//urls, err := getUrls()
	//showError(err, 40)
	//channel := make(chan int, 20)
	//for _, obj := range urls {
	//	channel <- 1
	//	go startDown(obj, channel)
	//}
	mergeFile("./storage/txt", "白夜宠物店 - 柒话.txt")
}

func getUrls() ([]Urls, error) {
	var contentJsonResult volumeStruct
	fp, err := os.Open("./list.json")
	if err != nil {
		return []Urls{}, err
	}
	defer fp.Close()
	fileInfo, err := fp.Stat()
	if err != nil {
		return []Urls{}, err
	}
	fileSize := fileInfo.Size()
	fileByte := make([]byte, fileSize)
	_, err = fp.Read(fileByte)

	data := handleChar(fileByte)
	if err != nil {
		return []Urls{}, err
	}

	err = json.Unmarshal(data, &contentJsonResult)
	if err != nil {
		return []Urls{}, err
	}
	var urls []Urls
	for index, volumn := range contentJsonResult.List {
		fileArr := strings.Split(volumn.ChapterName, "章")
		urls = append(urls,
			Urls{
				url:      fmt.Sprintf("https://www.yxlmdl.net/files/article/html555/8/8477/%s.html", volumn.ChapterId),
				fileName: fmt.Sprintf("第%03d章 %s", index+1, fileArr[1]),
			},
		)
	}

	return urls, nil
}

func handleChar(data []byte) []byte {
	data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
	data = bytes.Replace(data, []byte(" "), []byte(""), -1)
	data = bytes.Replace(data, []byte("\n"), []byte(""), -1)
	return data
}

func startDown(obj Urls, channel chan int) {
	content, err := getContent(obj.url)
	showError(err, 200)
	fileName := "storage/txt/" + obj.fileName + ".txt"
	mkdir(fileName)
	err = writeFile(content, fileName)
	showError(err, 210)
	fmt.Println("down text " + obj.fileName + " success~")
	<-channel
}

func getContent(url string) (string, error) {
	type header struct {
		key   string
		value string
	}
	var headers []header
	headers = []header{
		{
			key:   "Cookie",
			value: "Hm_lvt_f8405a99c4cd1af9524f2c4b916adcd8=1650181046; Hm_lvt_f5e665f36972b73374fde01d5178d046=1650181046; Hm_lvt_92aba3e4f105050c3554fc4ac3120577=1650181046; Hm_lpvt_f8405a99c4cd1af9524f2c4b916adcd8=1650188076; Hm_lpvt_92aba3e4f105050c3554fc4ac3120577=1650188076; Hm_lpvt_f5e665f36972b73374fde01d5178d046=1650188076",
		},
		{
			key:   "DNT",
			value: "1",
		},
		{
			key:   "sec-ch-ua",
			value: "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"100\", \"Google Chrome\";v=\"100\"",
		},
		{
			key:   "sec-ch-ua-mobile",
			value: "?0",
		},
		{
			key:   "sec-ch-ua-platform",
			value: "macOS",
		},
		{
			key:   "Sec-Fetch-Dest",
			value: "cors",
		},
		{
			key:   "Sec-Fetch-Site",
			value: "same-origin",
		},
		{
			key:   "User-Agent",
			value: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.60 Safari/537.36",
		},
		{
			key:   "X-Requested-With",
			value: "XMLHttpRequest",
		},
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	for _, value := range headers {
		req.Header.Add(value.key, value.value)
	}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	reader := transform.NewReader(response.Body, simplifiedchinese.GBK.NewDecoder())
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	data := string(body)
	dataArr := strings.Split(data, "cctxt=cctxt.replace(")
	data = dataArr[0]

	type replaceStruct struct {
		old string
		new string
	}

	var replaceArr []replaceStruct
	replaceArr = []replaceStruct{
		{"var cctxt=", ""},
		{"&nbsp;", " "},
		{"\n", ""},
		{"<br><br>", "\n"},
	}

	for _, reg := range dataArr[1:] {
		str := strings.Replace(reg, ");", "", -1)
		tmpArr := strings.Split(str, ",")
		replaceArr = append(replaceArr, replaceStruct{
			old: strings.Replace(strings.Replace(tmpArr[0], "/", "", 1), "/g", "", 1),
			new: strings.TrimLeft(strings.Replace(tmpArr[1], string([]byte{39, 13, 10}), "", 1), "'"),
		})
	}
	for _, replaceObj := range replaceArr {
		data = strings.Replace(data, replaceObj.old, replaceObj.new, -1)
	}
	data = strings.TrimLeft(data, string([]byte{39}))
	data = strings.TrimRight(data, string([]byte{39, 59, 13}))

	return data, nil
}

func writeFile(content string, file string) error {

	if err := os.WriteFile(file, []byte(content), 0666); err != nil {
		return err
	}
	return nil
}

func mkdir(path string) {
	_, err := os.Stat(path)
	if os.IsExist(err) == false {
		pathArr := strings.Split(path, "/")
		dir := strings.Join(pathArr[:len(pathArr)-1], "/")
		err = os.MkdirAll(dir, 0776)
		showError(err, 300)
		_, err = os.Create(path)
		showError(err, 310)
	}
}

func showError(err error, code int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(code)
	}
}

func mergeFile(path string, fileName string) {
	files, err := os.ReadDir(path)
	showError(err, 100)
	var filesArr []string
	for _, file := range files {
		filesArr = append(filesArr, path+"/"+file.Name())
	}
	sort.Strings(filesArr)
	lastFile := "storage/merge/" + fileName
	mkdir(lastFile)
	fp, err := os.OpenFile(lastFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0775)
	showError(err, 200)
	defer fp.Close()
	write := bufio.NewWriter(fp)
	for _, file := range filesArr {
		contentByte, err := os.ReadFile(file)
		showError(err, 260)
		fileArr := strings.Split(file, "/")
		fileName := strings.Replace(fileArr[len(fileArr)-1:][0], ".txt", "", 1)
		tmp := fileName + "\n" + string(contentByte) + "\n\n"
		_, err = write.WriteString(tmp)
		showError(err, 270)
	}
	write.Flush()
}
