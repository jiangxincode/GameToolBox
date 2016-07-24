// http proxy
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func HelloServer(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r)
	r.RequestURI = ""
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		fmt.Println("client :", err)
		w.Write([]byte("主机未找到"))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("body :", err)
	}
	w.Write(body)
	resp.Body.Close()
	fmt.Println("成功完成一次代理服务;url:", r.URL)
}

func main() {
	http.HandleFunc("/", HelloServer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
