package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var ch chan int = make(chan int)

func main() {
	sw := 6
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
	close(ch)
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
	ch <- 1
}
func printFile(url, fname string) {
	str, _ := netRequest("get", url)
	//ioutil.WriteFile(fname, []byte(str), 0666)
	file, _ := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	fwr := bufio.NewWriter(file)
	fwr.Write([]byte(str))
	fwr.Flush()
	ch <- 1
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
	for j := 1; j <= 100; {
		j += <-ch
	}
	//time.Sleep(time.Second * 3)
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
	for j := 1; j <= 100; {
		j += <-ch
	}
	//time.Sleep(time.Second * 5)
}

type workbchDB struct {
	DBname, user, password, host string
	link                         *sql.DB
}

func (wDB *workbchDB) connectMySQL() (*sql.DB, error) {
	driverName := "mysql"
	connectString := fmt.Sprintf("%s:%s@tcp(%s)/%s", wDB.user, wDB.password, wDB.host, wDB.DBname)
	db, err := sql.Open(driverName, connectString)
	if err != nil {
		log.Fatal(err)
	}
	wDB.link = db
	return db, err
}
func (wdb *workbchDB) createTable(tname string, s interface{}) {
	/* stmt, err := db.Prepare(createTable)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec()
		if err != nil {
			panic(err)
		}
	}

	var createTable = `
	CREATE TABLE IF NOT EXISTS people (
	     user_id      INTEGER PRIMARY KEY AUTO_INCREMENT
	    ,username     VARCHAR(32)
	    ,phone        VARCHAR(32)
	); */
}
func (wdb *workbchDB) dropTable(tname string) bool {
	dropT := fmt.Sprintf("DROP TABLE %s;", tname)
	stat, err := wdb.link.Prepare(dropT)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer stat.Close()
	_, err = stat.Exec()
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

type Post struct {
	UserId int
	Id     int
	Title  string
	Body   string
}
type Comment struct {
	PostId int
	Id     int
	Name   string
	Email  string
	Body   string
}
type ElemReq interface {
	writeDB()
}

func (p *Post) writeDB() {
	fmt.Println(p.Id)
}
func (c *Comment) writeDB() {
	fmt.Println(c.Id)
}

func task6() {
	/*
		   получить json posts?userId=7  обработать: получить структуру постов
		   в горутины 1го ур(анонимные функции) выдавать по элементу(посту) структуры
		   		в этой рутине произвести запись поста в БД, получить json comments?postId обработать: получить структуру комментов
		   		в горутины 2го ур выдавать по элементу(комменту) структуры
				   в этой рутине произвести запись коммента
	*/
	wdb := workbchDB{DBname: "edudb", user: "utest", password: "12345", host: "localhost:3306"}
	db, _ := wdb.connectMySQL()
	defer db.Close()
	db
	url := netResource + "posts?userId=7"
	resp, _ := netRequest("get", url)
	var posts []Post
	err := json.Unmarshal([]byte(resp), &posts)
	if err != nil {
		log.Fatal(err)
	}
	var ch1 chan int = make(chan int)
	for _, value := range posts {
		go procPost(&wdb, value, ch1)
	}
	for i := 1; i <= len(posts); {
		i += <-ch1
	}
	close(ch1)
}

func write2DB(wdb *workbchDB, e ElemReq) {
	//использовать мьютексы во время инсертов

}

func procPost(wdb *workbchDB, p Post, ch chan int) {
	url := netResource + "comments?postId=" + strconv.Itoa(p.Id)
	resp, _ := netRequest("get", url)
	var comments []Comment
	err := json.Unmarshal([]byte(resp), &comments)
	if err != nil {
		log.Fatal(err)
	}
	var ch2 chan int = make(chan int)
	for _, value := range comments {
		go procComment(wdb, value, ch2)
	}
	for i := 1; i <= len(comments); {
		i += <-ch2
	}
	close(ch2)
	ch <- 1
}

func procComment(wdb *workbchDB, c Comment, ch chan int) {
	fmt.Println(c)
	ch <- 1
}
