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
	"sync"

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
	mux                          sync.Mutex
}

func (wdb *workbchDB) connectMySQL() (*sql.DB, error) {
	driverName := "mysql"
	connectString := fmt.Sprintf("%s:%s@tcp(%s)/%s", wdb.user, wdb.password, wdb.host, wdb.DBname)
	db, err := sql.Open(driverName, connectString)
	if err != nil {
		log.Fatal(err)
	}
	wdb.link = db
	return db, err
}

type tableStruct struct {
	tname     string
	colStruct map[string]string
}

func (wdb *workbchDB) createTable(ts tableStruct) {
	createT := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", ts.tname)
	for colN, colT := range ts.colStruct {
		tmpstr := fmt.Sprintf("`%s` %s,", colN, colT)
		createT += tmpstr
	}
	createT = createT[:len(createT)-1]
	createT += ");"
	stat, err := wdb.link.Prepare(createT)
	if err != nil {
		log.Fatal(err)
	}
	defer stat.Close()

	_, err = stat.Exec()
	if err != nil {
		log.Fatal(err)
	}
}
func (wdb *workbchDB) dropTable(tname string) bool {
	dropT := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tname)
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

type post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}
type comment struct {
	PostID int    `json:"postId"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}
type elemReq interface {
	writeDB(wdb *workbchDB)
}

func (p *post) writeDB(wdb *workbchDB) {
	fmt.Println(p.ID)
	sql := fmt.Sprintf(`INSERT INTO posts(UserId, Id, Title, Body) VALUES(%d , %d, "%s", "%s");`, p.UserID, p.ID, p.Title, p.Body)
	stat, err := wdb.link.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer stat.Close()
	wdb.mux.Lock()
	_, err = stat.Exec()
	wdb.mux.Unlock()
	if err != nil {
		log.Fatal(err)
	}
}
func (c *comment) writeDB(wdb *workbchDB) {
	fmt.Println(c.ID)
	sql := fmt.Sprintf(`INSERT INTO comments(PostId, Id, Name, Email,Body) VALUES(%d , %d, "%s", "%s", "%s");`, c.PostID, c.ID, c.Name, c.Email, c.Body)
	stat, err := wdb.link.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer stat.Close()
	wdb.mux.Lock()
	_, err = stat.Exec()
	wdb.mux.Unlock()
	if err != nil {
		log.Fatal(err)
	}
}

func task6() {
	wdb := workbchDB{DBname: "edudb", user: "utest", password: "12345", host: "localhost:3306"}
	db, _ := wdb.connectMySQL()
	defer db.Close()
	columnsPosts := map[string]string{
		"userId": "INT",
		"id":     "INT PRIMARY KEY",
		"title":  "VARCHAR(255)",
		"body":   "VARCHAR(255)",
	}
	var tsPosts = tableStruct{"posts", columnsPosts}
	wdb.dropTable("posts")
	wdb.createTable(tsPosts)
	columnsComments := map[string]string{
		"postId": "INT",
		"id":     "INT PRIMARY KEY",
		"name":   "VARCHAR(255)",
		"email":  "VARCHAR(255)",
		"body":   "VARCHAR(255)",
	}
	var tsComments = tableStruct{"comments", columnsComments}
	wdb.dropTable("comments")
	wdb.createTable(tsComments)
	url := netResource + "posts?userId=7"
	resp, _ := netRequest("get", url)
	var posts []post
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

func write2DB(wdb *workbchDB, e elemReq) {
	e.writeDB(wdb)
}

func procPost(wdb *workbchDB, p post, ch chan int) {
	var e elemReq
	e = &p
	write2DB(wdb, e)
	url := netResource + "comments?postId=" + strconv.Itoa(p.ID)
	resp, _ := netRequest("get", url)
	var comments []comment
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

func procComment(wdb *workbchDB, c comment, ch chan int) {
	var e elemReq
	e = &c
	write2DB(wdb, e)
	ch <- 1
}
