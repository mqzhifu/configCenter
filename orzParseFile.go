package configcenter

import (
	"bufio"
	"io"
	"os"
	"strings"
	"encoding/json"
	"src/zlib"
)
const DefaultCommentChar = ";"

type LoadIniFile struct {
	pathFile		string
	fileSize		int64
	contentSection 	map[string]map[string]string
	contentNormal 	map[string]string
	content			map[string]interface{}
}

func (loadIniFile *LoadIniFile)getContent(){
	json.Marshal(loadIniFile.content)
}
//设置文件路径
func (loadIniFile *LoadIniFile)setPathFile(pathFile string)error{
	zlib.MyPrint(" setPathFile ",pathFile)
	loadIniFile.pathFile = pathFile
	info, err := os.Stat(pathFile)
	if err != nil {
		zlib.NewCoder(603,err.Error())
	}

	if info.IsDir() {
		return  zlib.NewCoder(601,"file path is a folder.")
	}

	if info.Size() == 0 {
		return  zlib.NewCoder(602,"file is empty.")
	}
	loadIniFile.fileSize = info.Size()

	//if os.IsNotExist(err) {
	//	return sectionContent,normalContent,errors.New("file IsNotExist")
	//}
	return nil
}
//开始解析一个配置文件
func (loadIniFile *LoadIniFile)process()error{
	//fmt.Println("loadIniFile process",loadIniFile.pathFile)
	if loadIniFile.pathFile == "" {
		return zlib.NewCoder(600,"config path is empty")
	}

	fileFd, err := os.Open(loadIniFile.pathFile)
	if err != nil {
		//return sectionContent,normalContent,err
		return  zlib.NewCoder(604,err.Error())
	}
	defer fileFd.Close()

	sectionContent := make(map[string]map[string]string)
	normalContent := make(map[string]string)

	reader := bufio.NewReader(fileFd)
	lastSection := ""

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		lineStr := string(line)

		exist, first, start := processComment(lineStr)
		//whole line is comment
		if exist && first {
			continue
		}

		section, sectionExist := processSection(lineStr, start)
		//whole line is section
		if sectionExist {
			_, ok := sectionContent[section]
			//section already exist
			if ok {
				continue
			}
			sectionContent[section] = make(map[string]string)
			lastSection = section
		}

		key, value, ok := processLine(lineStr, start)
		if ok {
			//这里有个bug , 一个配置kv 即存了 normal 又存了 section
			//normalContent[key] = value
			//dont have section
			if lastSection == "" {
				continue
			}
			sectionContent[lastSection][key] = value
		}
	}
	sectionContentLen := len(sectionContent)
	normalContentLen := len(normalContent)
	if sectionContentLen == 0 && normalContentLen == 0 {
		return  zlib.NewCoder(607,"file is empty.")
	}
	//fmt.Println("sectionContentLen",sectionContentLen,"normalContentLen",normalContentLen)
	loadIniFile.contentNormal = normalContent
	loadIniFile.contentSection = sectionContent
	loadIniFile.mergeContent()
	zlib.MyPrint("load file ,size : ",loadIniFile.fileSize," byte , contentSection:" ,sectionContent , " content : ",loadIniFile.content)
	return nil
}

func (loadIniFile *LoadIniFile) search(path string)interface{}{
	zlib.MyPrint("search file content: ",path)
	path = strings.TrimSpace(path)//过滤首尾空格
	path = strings.Trim(path,"/")//过滤首尾反斜杠
	pathSplit := strings.Split(path, "/")
	if len(pathSplit) > 2{
		return ""
	}
	if len(pathSplit) == 1{
		for k, v := range loadIniFile.content {
			if k == pathSplit[0]{
				return v
			}
		}
	}else if len(pathSplit) == 2{
		//zlib.MyPrint("sub : ",pathSplit[1])
		for k, v := range loadIniFile.content {
			if k == pathSplit[0]{
				subLevel := v.(map[string]string)
				for k2, v2 := range subLevel {
					if k2 == pathSplit[1] {
						return v2
					}
				}
			}
		}
	}
	return ""

}
//将：普通配置 和 模块配置 两种配置内容，合并为一个大的map
func (loadIniFile *LoadIniFile)mergeContent(){
	var content map[string]interface{}
	content = make(map[string]interface{})

	for key, value := range loadIniFile.contentNormal {
		content[key] = value
	}
	if len(loadIniFile.contentSection) > 0 {
		for key, value := range loadIniFile.contentSection {
			content[key] = value
		}
	}

	loadIniFile.content = content
}

func processLine(context string, commentStart int) (key string, value string, ok bool) {
	start := strings.Index(context, "=")
	if start != -1 && start != 0 {
		//comment exist
		if commentStart != -1 {
			if commentStart < start {
				return "", "", false
			}
		}
		key = string([]rune(context)[:start])
		key = strings.Replace(key, " ", "", -1)
		if key == "" {
			return "", "", false
		}
		value = string([]rune(context)[start+1:])
		value = strings.Replace(value, " ", "", -1)
		if value == "" {
			return key, "", true
		}
		//check comment in value
		commentStart2 := strings.Index(value, DefaultCommentChar)
		if commentStart2 != -1 && commentStart2 != 0 { //comment in middle of value
			value = string([]rune(value)[:commentStart2])
			return key, value, true
		} else if commentStart2 == 0 { //comment at first of value
			return key, "", true
		} else { //comment not exist
			return key, value, true
		}
	} else {
		return "", "", false
	}
}

func processSection(context string, commentStart int) (str string, exist bool) {
	section := ""
	start := strings.Index(context, "[")
	end := strings.Index(context, "]")
	if start != -1 && end != -1 {
		//comment exist
		if commentStart != -1 {
			if commentStart < start {
				return section, false
			}
		}
		section = string([]rune(context)[start+1 : end])
		return section, true
	} else {
		return section, false
	}
}

func processComment(context string) (exist bool, isFirst bool, index int) {
	start := strings.Index(context, DefaultCommentChar)
	if start == 0 {
		return true, true, start
	} else if start == -1 {
		return false, false, start
	} else {
		return true, false, start
	}
}
