package configcenter

import (
	"io/ioutil"
	"src/zlib"
	"strings"
	"fmt"
	"strconv"
	"os"
	"encoding/json"
)
type Configer struct {
	//配置文件加载到内存，占最大值,单位:mb
	FileTotalSizeMax int
	//单文件最大值
	FileSizeMax	int
	//最多加载多少个文件
	FileCntMax int
	//加载哪一个目录
	RootPath string
	//加载哪些-配置文件扩展名
	AllowExtType string
	//所有配置文件的内容-容器 | allConfig[appName][fileName]
	ContainerFileContent map[string]map[string]LoadIniFile
	//路径最多，支持几级，目前是4级
	PathLevel int
}
//创建一个 类，并初始化成员变量
func NewConfiger (fileTotalSizeMax int ,fileSizeMax int , fileCntMax int ,allowExtType string) *Configer{

	configer := new (Configer)
	configer.FileTotalSizeMax = fileTotalSizeMax
	configer.FileSizeMax = fileSizeMax
	configer.FileCntMax = fileCntMax
	configer.AllowExtType = allowExtType
	configer.PathLevel = 4

	zlib.MyPrint("create Configer obj , fileTotalSizeMax: ",fileTotalSizeMax , "m , fileSizeMax:",fileSizeMax ,"m ,fileCntMax:",fileCntMax ," , allowExtType:",allowExtType)

	return configer
}
//检查 成员变量 初始值 是否正确
func (configer *Configer) checkMemberVariable()error{
	if configer.FileTotalSizeMax == 0 {
		return zlib.NewCoder(400,"FileTotalSizeMax == 0")
	}

	if configer.FileSizeMax == 0 {
		return zlib.NewCoder(401,"FileSizeMax == 0")
	}

	if configer.FileCntMax == 0 {
		return zlib.NewCoder(402,"FileCntMax == 0")
	}

	if configer.RootPath == "" {
		return zlib.NewCoder(403,"RootPath is empty or not exec <StartLoading>")
	}
	return nil
}
//开始加载 - 指定目录 - 所有ini配置文件到内存
func (configer *Configer) StartLoading(rootPath string)error{
	zlib.MyPrint("StartLoading init..."+rootPath)
	isDir := checkIsDir(rootPath)
	if !isDir{
		return zlib.NewCoder(300,"path not is dir")
	}

	if configer.RootPath != ""{
		return zlib.NewCoder(301,"configer.RootPath has been value")
	}
	configer.RootPath = rootPath
	err := configer.checkMemberVariable()
	if err != nil{
		newErr := zlib.Wrap(err,"StartLoading checkMemberVariable")
		return newErr
	}
	zlib.MyPrint("check path ok~")
	//获取指定文件夹下面的，所有目录及文件
	fileList,dirList,err := configer.GetAllFiles(rootPath)
	if err != nil{
		WrapErr := zlib.Wrap(err,"GetAllFiles")
		return WrapErr
	}
	//fmt.Println("dirList",dirList)
	if len(dirList) == 0 {
		return zlib.NewCoder(700,"this dir ,no include app dir")
	}
	//fmt.Println("fileList",fileList)
	if len(fileList) <= 0 {
		return zlib.NewCoder(701,"this dir ,no include any files.")
	}

	if len(fileList) > configer.FileCntMax {
		return zlib.NewCoder(702,"fileCntMax:"+strconv.Itoa(configer.FileCntMax))
	}
	zlib.MyPrint("get path dir and file ok,files num :",len(fileList)," dirs num : ",len(dirList))
	//初始化全局-内存池，用于接收文件内容
	containerFileContent := make(map[string]map[string]LoadIniFile)
	for i:=0;i<len(dirList);i++{
		appName := configer.GetAppNameByPath(dirList[i],rootPath)
		configFile := make(map[string]LoadIniFile)
		containerFileContent[appName] = configFile
	}
	//汇总所有文件内容的总大小
	var fileTotalSize int64 = 0
	//开始循环读取每个文件的内容
	for i := 0;i<len(fileList);i++{
		//定义一个，解析配置文件的，类/结构体
		var loadIniFile LoadIniFile
		//告诉  解析类  解析哪个文件
		err := loadIniFile.setPathFile(fileList[i])
		if err != nil {
			newErr := zlib.Wrap(err,"loadIniFile setPathFile")
			return newErr
		}
		//解析文件类，开始处理
		err2 := loadIniFile.process()
		if err2 != nil {
			newErr := zlib.Wrap(err,"LoadIniFileContent")
			return newErr
		}
		//fmt.Println(loadIniFile.contentSection)
		//判断 单个文件内容最大值
		//fmt.Println("file size : ",loadIniFile.fileSize)
		if loadIniFile.fileSize > int64(configer.FileSizeMax * 1024 * 1024)  {
			return zlib.NewCoder(703,"fileSizeMax:"+strconv.Itoa(configer.FileSizeMax))
		}
		//判断 所有文件内容 最大值
		fileTotalSize += loadIniFile.fileSize
		if fileTotalSize > int64(configer.FileTotalSizeMax * 1024 * 1024)  {
			return zlib.NewCoder(704,"fileTotalSize:"+strconv.Itoa(configer.FileTotalSizeMax))
		}
		//构建key :  文件夹名(appName)/文件名(不包括扩展名)/元素-key/元素-value
		appName := configer.GetAppNameByPathFile(fileList[i],rootPath)
		fileName ,_ := configer.GetFileNameByPath(fileList[i],rootPath)

		containerFileContent[appName][fileName] = loadIniFile
	}
	zlib.MyPrint("fileTotalSize : ",fileTotalSize,"byte")
	//fmt.Println(containerFileContent)
	//将 新文件解析后的内容，放置到统一的map容器中
	configer.ContainerFileContent = containerFileContent
	//configer.ShowFormatContent()
	return nil
}
//检查一个 字符串 是否为目录
func checkIsDir(path string)bool{
	s,err:=os.Stat(path)
	if err!=nil{
		return false
	}
	return s.IsDir()
}
//过滤掉绝对路径中的根目录，取出相对路径
//给定两个值： /xxxxx/文件夹/ 和  /xxxxxx  ，把/xxxxxx 删除
func (configer *Configer) GetAppNameByPath (path  string , rootPath string) string{
	return strings.Replace(path, rootPath+"/","",-1)
}
//给定  /xxxxx/文件夹/文件名.ini ，取出最后一个目录的名称
func (configer *Configer) GetAppNameByPathFile (pathFile  string , rootPath string)string{
	//先删除绝对路径中的根目录，取出相对路径
	relativePath := configer.GetAppNameByPath(pathFile,rootPath)
	relativePath = strings.Trim(relativePath,"/")
	relativePathSplit := strings.Split(relativePath, "/")
	appName := relativePathSplit[0]
	//fmt.Println("relativePath",relativePath,appName)
	return appName
}
//给定一个完整路径，取出该路径中，文件名及扩展名
func (configer *Configer) GetFileNameByPath(path string , rootPath string)(fileName string ,extName string){
	fileNameSplit := strings.Split(path, "/")
	//取最后一个元素，文件名+扩展名
	fileNameAndExt :=  fileNameSplit[len(fileNameSplit) - 1]
	fileNameAndExtSplit :=  strings.Split(fileNameAndExt, ".")
	//取出文件名，不包含扩展名
	fileName = fileNameAndExtSplit[0]
	extName = fileNameAndExtSplit[1]

	//fmt.Println(fileNameSplit,fileNameAndExt,fileNameAndExtSplit,fileName,extName)
	return fileName,extName
}
//搜索
func (configer *Configer) Search(path string)(data string,err error){
	zlib.MyPrint(" search : ",path)
	data  = "{}" //返回一个空对象，用于默认值
	if path == ""{
		return data,zlib.NewCoder(400,"path is empty")
	}
	err2 := configer.checkMemberVariable()
	if err2 != nil{
		newErr := zlib.Wrap(err2,"Search checkMemberVariable")
		return data,newErr
	}

	path = strings.TrimSpace(path)//过滤首尾空格
	path = strings.Trim(path,"/")//过滤首尾反斜杠
	pathSplit := strings.Split(path, "/")
	if len(pathSplit) > configer.PathLevel {
		return data,zlib.NewCoder(401,"目前路径，仅支持最大4级   项目名/文件名/模块名/key")
	}
	appName := pathSplit[0]
	//先搜索一级路径 : appName
 	app, ok := configer.ContainerFileContent[appName]
 	if !ok {
 		fmt.Print(path," path appName not in mem")
		return data,nil
	}
	if len(pathSplit) == 1{
		var result map[string]interface{}
		result = make(map[string]interface{})
		for fileName, fileContent := range app {
			result[fileName] = fileContent.content
			//fmt.Println(fileContent.content)
		}
		jsonRs , _ := json.Marshal(result)
		fmt.Println(string(jsonRs))
		return string(jsonRs),nil
	}else if len(pathSplit) == 2{
		fileName := pathSplit[1]
		fileContent, ok := configer.ContainerFileContent[appName][fileName]
		if !ok {
			//fmt.Print(path,"path file not in mem")
			return data,nil
		}
		jsonRs , _ := json.Marshal(fileContent.content)
		fmt.Println(string(jsonRs))
		return string(jsonRs),nil
	}else{
		fileName := pathSplit[1]
		fileContent, ok := configer.ContainerFileContent[appName][fileName]
		if !ok {
			//fmt.Print(path,"path file not in mem")
			return data,nil
		}

		appNameFileName := appName+"/"+fileName
		searchFileContentPath := strings.Replace(path, appNameFileName,"",-1)
		//zlib.MyPrint("searchFileContentPath : ",searchFileContentPath)
		content := fileContent.search(searchFileContentPath)
		//zlib.MyPrint(content)
		jsonRs , _ := json.Marshal(content)
		return string(jsonRs),nil
	}
}
//这里先预留吧，应该是个 循环递归遍历搜索
func (configer *Configer)  SearchContainerFileContent(){

}
//遍历 内存中的 所有加载的内容 ,可视化-查看
func (configer *Configer) ShowFormatContent()error{
	err2 := configer.checkMemberVariable()
	if err2 != nil{
		newErr := zlib.Wrap(err2,"Search checkMemberVariable")
		return  newErr
	}

	for appName, fileContent := range configer.ContainerFileContent {
		fmt.Println("appName : " + appName)
		for fileName, LoadIniFile := range fileContent {
			fmt.Println("  file name : " + fileName)
			fmt.Println("    contentSection : ",LoadIniFile.contentSection)
			fmt.Println("    contentNormal  : ",LoadIniFile.contentNormal)
		}
	}
	return nil;
}
//获取指定目录下的所有文件,包含子目录下的文件
func  (configer *Configer) GetAllFiles(dirPth string) (files []string, returnDirs []string ,err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return files,returnDirs,zlib.NewCoder(600,err.Error())
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			returnDirs = append(returnDirs, dirPth+PthSep+fi.Name())
			configer.GetAllFiles(dirPth + PthSep + fi.Name())
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), configer.AllowExtType)
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _,_ := configer.GetAllFiles(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, returnDirs,nil
}

