package configcenter

import "src/zlib"

func StartHttpd(){
	zlib.LogLevelFlag = zlib.LOG_LEVEL_DEBUG
	iniConfigDirPath := "/data/www/golang/src/configcenter"
	host := "0.0.0.0"
	port := 1234

	configer := Configer{
		FileTotalSizeMax : 100,
		FileSizeMax : 2,
		FileCntMax : 100,
		RootPath : iniConfigDirPath,
		AllowExtType : "ini",
	}
	//my.StartLoading(iniConfigDirPath)
	//systemConfigStr,_ := configCenter.Search("system")
	httpd := NewHttpd(port,host,configer)
	httpd.Start()
}
