package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// 这一步是去掉静态文件目录的前缀 static，使得文件系统根路径就是 static 里内容
	staticContent, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	// 静态文件 embed.FS
	fs := http.FileServerFS(staticContent)

	// 将根路径 "/" 的请求指向文件服务器
	// http.Handle("/", fs)

	// rootHandler 来处理根路径 "/" 的请求
	http.HandleFunc("/", rootHandler(fs))

	// API 路由
	http.HandleFunc("/api/okhk", apiHandler)

	port := "8080"
	log.Printf("服务启动，监听端口: %s", port)
	log.Printf("请在浏览器中打开 http://localhost:%s", port)

	// 启动HTTP服务
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("服务启动失败: %s\n", err)
	}
}
