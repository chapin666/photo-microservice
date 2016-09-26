/**
*
* 存储图片微服务，然后使用键值界面和参数表明图片是修改前还是修改后。
* 为了简洁起见，我们把图片保存在一个以图片状态（完成／进行中）命名的文件夹。
*
**/
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func main() {

	if !registerInKVStore() {
		return
	}

	http.HandleFunc("/sendImage", receiveImage)
	http.HandleFunc("/getImage", serveImage)
	http.ListenAndServe(":3002", nil)
}

// register service to keyvalue
func registerInKVStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return false
	}

	storgeAddress := os.Args[1]
	keyValueStoreAddress := os.Args[2]

	response, err := http.Post("http://"+keyValueStoreAddress+"/set?key=storgeAddress&value="+storgeAddress, "", nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error:", string(data))
		return false
	}

	return true
}

func receiveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", "Wrong input id.")
			return
		}
		if values.Get("state") != "working" && values.Get("state") != "finished" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", "Wrong input state.")
			return
		}
		_, err = strconv.Atoi(values.Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", "Wrong input id.")
			return
		}

		file, err := os.Create("/tmp/" + values.Get("state") + "/" + values.Get("id") + ".png")
		defer file.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}

		_, err = io.Copy(file, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", err)
			return
		}
		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted.")
	}
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", "Wrong input id.")
			return
		}

		file, err := os.Open("/tmp/" + values.Get("state") + "/" + values.Get("id") + ".png")
		defer file.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", err)
			return
		}

		_, err = io.Copy(w, file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted.")
	}
}
