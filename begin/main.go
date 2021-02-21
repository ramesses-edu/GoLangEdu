package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	sw := 1
	switch sw {
	case 1: // Hello, NIX Education
		task1()
	case 2: // repository created
	case 3: // work with Net
		task3()
	case 4: // work with gorootine
		task4()
	case 5: //work with filesys
		task5()
	case 6: // work with DB
		task6()
	}
}

var netResource string = "https://jsonplaceholder.typicode.com/"

func netRequest(metod, url string) (resp string, err error) {
	resp = ""
	err = nil
	switch metod {
	case "get":
		rsp, er := http.Get(url)
		if er != nil {
			log.Fatal(er)
		}
		rspb, er := ioutil.ReadAll(rsp.Body)
		rsp.Body.Close()
		if er != nil {
			log.Fatal(er)
		}
		resp = string(rspb)
		err = er
	case "post":
	case "put":
	case "patch":
	case "delete":
	}
	return resp, err
}

func printScr(url string) {
	str, _ := netRequest("get", url)
	fmt.Print(str)
}
func printFile(url, fname string) {
	str, _ := netRequest("get", url)
	//ioutil.WriteFile(fname, []byte(str), 0666)
	file, _ := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	fwr := bufio.NewWriter(file)
	fwr.Write([]byte(str))
	fwr.Flush()
}

func task1() {
	fmt.Println("Hello, NIX Education")
}

func task3() {
	page := "posts/"
	url := netResource + page
	printScr(url)
}

func task4() {
	page := "posts/"
	url := ""
	for i := 1; i <= 100; i++ {
		url = netResource + page + strconv.Itoa(i)
		go printScr(url)
	}
	time.Sleep(time.Second * 3)
}

func task5() {
	path := "./storage/posts/"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0666)
	} else if _, err := os.Stat(path); err == nil {
		os.RemoveAll(path)
		os.MkdirAll(path, 0666)
	}
	page := "posts/"
	url := ""
	for i := 1; i <= 100; i++ {
		url = netResource + page + strconv.Itoa(i)
		fname := path + strconv.Itoa(i) + ".txt"
		go printFile(url, fname)
	}
	time.Sleep(time.Second * 5)
}

func task6() {

}
