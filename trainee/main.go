package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	netResource string = "https://jsonplaceholder.typicode.com/"
	userDB      string = "utest"
	passDB      string = "12345"
	hostDB      string = "localhost"
	portDB      string = "3306"
	nameDB      string = "edudb"
	mux         sync.Mutex
)

type user struct {
	User     xml.Name `xml:"User" json:"-" gorm:"-"`
	ID       int      `xml:"Id" json:"id" gorm:"column:id;primaryKey"`
	Name     string
	Username string
	Email    string
	Address  address `gorm:"-"`
	Phone    string
	Website  string
	Company  company `gorm:"-"`
}
type address struct {
	Address xml.Name `xml:"Address" json:"-" gorm:"-"`
	ID      int      `xml:"-" json:"-" gorm:"primaryKey"`
	UserID  int      `xml:"-" json:"-" gorm:"column:userId"`
	Street  string
	Suite   string
	City    string
	Zipcode string
	Geo     geo `gorm:"-"`
}
type geo struct {
	Geo       xml.Name `xml:"Geo" json:"-" gorm:"-"`
	ID        int      `xml:"-" json:"-" gorm:"primaryKey"`
	AddressID int      `xml:"-" json:"-" gorm:"column:addressId"`
	Lat       float32
	Lng       float32
}
type company struct {
	Company     xml.Name `xml:"Company" json:"-" gorm:"-"`
	Name        string
	CatchPhrase string
	Bs          string
	ID          int `xml:"-" json:"-" gorm:"primaryKey"`
	UserID      int `xml:"-" json:"-" gorm:"column:userId"`
}

///////////////////////////////////////////////////////////////////////////
type post struct {
	Post   xml.Name `xml:"Post" json:"-" gorm:"-"`
	UserID int      `json:"userId" gorm:"column:userId"`
	ID     int      `json:"id" gorm:"column:id;primaryKey"`
	Title  string   `json:"title" gorm:"column:title;type:VARCHAR(256)"`
	Body   string   `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}
type comment struct {
	Comment xml.Name `xml:"Comment" json:"-" gorm:"-"`
	PostID  int      `json:"postId" gorm:"column:postId"`
	ID      int      `json:"id" gorm:"column:id;primaryKey"`
	Name    string   `json:"name" gorm:"column:name;type:VARCHAR(256)"`
	Email   string   `json:"email" gorm:"column:email;type:VARCHAR(256)"`
	Body    string   `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

func main() {
	handle := http.NewServeMux()
	handle.HandleFunc("/", handleFunc)
	srv := http.Server{
		Addr:    "localhost:80",
		Handler: handle,
	}
	srv.ListenAndServe()
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	//разбор урла
	//получение параметров r.FormValue()  or r.ParseForm() --> r.Form
	//подкл к БД
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", userDB, passDB, hostDB, portDB, nameDB)
	gormD := mysql.New(mysql.Config{
		DSN:               dsn,
		DefaultStringSize: 256,
	})
	db, err := gorm.Open(gormD, &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	sql, _ := db.DB()
	defer sql.Close()
	sql.Ping()
	if !db.Migrator().HasTable(&user{}) {
		db.Migrator().CreateTable(&user{})
	}
	if !db.Migrator().HasTable(&address{}) {
		db.Migrator().CreateTable(&address{})
	}
	if !db.Migrator().HasTable(&geo{}) {
		db.Migrator().CreateTable(&geo{})
	}
	if !db.Migrator().HasTable(&company{}) {
		db.Migrator().CreateTable(&company{})
	}
	////////////////////////////////////////////
	r.ParseForm()
	switch r.Method {
	case http.MethodGet:
		methodGet(w, r, db)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	}

}

func methodGet(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	formatXML := false
	for key := range r.Form {
		if strings.ToLower(key) == "xml" {
			formatXML = true
			break
		}
	}
	fmt.Fprintf(w, "Hello \n")
	fmt.Fprintf(w, "URI %q \n", r.RequestURI)
	fmt.Fprintf(w, "URL %q \n", r.URL)
	fmt.Fprintf(w, "Path %q \n", r.URL.Path)
	fmt.Fprintf(w, "formatXML %t", formatXML)
	//var p user
	//getFromDB(db, p, map[string]interface{}{"ID": 1})
}

//
//map с параметрами для gorm
func getFromDB(db *gorm.DB, obj interface{}, param map[string]interface{}) interface{} {
	typeObj := reflect.TypeOf(obj)
	switch strings.ToLower(typeObj.Name()) {
	case "user":
		var users []user
		db.Where(param).Find(&users)
		for key, value := range users {
			var userAdrr address
			result := db.Where("userId = ?", strconv.Itoa(value.ID)).Last(&userAdrr)
			if result.Error != gorm.ErrRecordNotFound {
				users[key].Address = userAdrr
				var addrGeo geo
				result := db.Where("addressId = ?", strconv.Itoa(userAdrr.ID)).Last(&addrGeo)
				if result.Error != gorm.ErrRecordNotFound {
					users[key].Address.Geo = addrGeo
				}
			}
			var userComp company
			result = db.Where("userId = ?", strconv.Itoa(value.ID)).Last(&userComp)
			if result.Error != gorm.ErrRecordNotFound {
				users[key].Company = userComp
			}
		}
		return users
	case "post":
		var posts []post
		db.Where(param).Find(&posts)
		return posts
	case "comment":
		var comments []comment
		db.Where(param).Find(&comments)
		return comments
	}
	return nil
}

/*
{
    "id": 1,
    "name": "Leanne Graham",
    "username": "Bret",
    "email": "Sincere@april.biz",
    "address": {
      "street": "Kulas Light",
      "suite": "Apt. 556",
      "city": "Gwenborough",
      "zipcode": "92998-3874",
      "geo": {
        "lat": "-37.3159",
        "lng": "81.1496"
      }
    },
    "phone": "1-770-736-8031 x56442",
    "website": "hildegard.org",
    "company": {
      "name": "Romaguera-Crona",
      "catchPhrase": "Multi-layered client-server neural-net",
      "bs": "harness real-time e-markets"
    }
  }
*/
