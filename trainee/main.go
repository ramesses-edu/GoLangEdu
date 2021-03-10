package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
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

type users struct {
	XMLName xml.Name `xml:"users" json:"-" gorm:"-"`
	Users   interface{}
}
type user struct {
	XMLName  xml.Name `xml:"user" json:"-" gorm:"-"`
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
	XMLName xml.Name `xml:"address" json:"-" gorm:"-"`
	ID      int      `xml:"-" json:"-" gorm:"primaryKey"`
	UserID  int      `xml:"-" json:"-" gorm:"column:userId"`
	Street  string
	Suite   string
	City    string
	Zipcode string
	Geo     geo `gorm:"-"`
}
type geo struct {
	XMLName   xml.Name `xml:"geo" json:"-" gorm:"-"`
	ID        int      `xml:"-" json:"-" gorm:"primaryKey"`
	AddressID int      `xml:"-" json:"-" gorm:"column:addressId"`
	Lat       float32
	Lng       float32
}
type company struct {
	XMLName     xml.Name `xml:"company" json:"-" gorm:"-"`
	Name        string
	CatchPhrase string
	Bs          string
	ID          int `xml:"-" json:"-" gorm:"primaryKey"`
	UserID      int `xml:"-" json:"-" gorm:"column:userId"`
}

///////////////////////////////////////////////////////////////////////////
type posts struct {
	XMLName xml.Name `xml:"posts" json:"-" gorm:"-"`
	Posts   interface{}
}
type post struct {
	XMLName xml.Name `xml:"post" json:"-" gorm:"-"`
	UserID  int      `json:"userId" gorm:"column:userId"`
	ID      int      `json:"id" gorm:"column:id;primaryKey"`
	Title   string   `json:"title" gorm:"column:title;type:VARCHAR(256)"`
	Body    string   `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

//////////////////////////////////////////////////////////////////////////
type comments struct {
	XMLName  xml.Name `xml:"comments" json:"-" gorm:"-"`
	Comments interface{}
}
type comment struct {
	XMLName xml.Name `xml:"comment" json:"-" gorm:"-"`
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
		//SELECT
		methodGet(w, r, db)
	case http.MethodPost:
		//INSERT
		metodPOST(w, r, db)
	case http.MethodPut:
		//UPDATE
	case http.MethodDelete:
		//DELETE
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
	reObj := regexp.MustCompile(`^(\/[a-zA-Z]+)(\/)??$`)     // любой /aaaaa или /aaaaa/
	reID := regexp.MustCompile(`^(\/[a-zA-Z]+\/)\d+(\/)??$`) // любой /aaaaa/111  или /aaaaa/11111/
	reSymb := regexp.MustCompile(`[a-zA-Z]+`)
	reNum := regexp.MustCompile(`[0-9]+`)

	var mode string
	var id int = 0
	if reID.Match([]byte(r.URL.Path)) {
		mode = reSymb.FindString(r.URL.Path)
		idStr := reNum.FindString(r.URL.Path)
		if idStr != "" {
			id, _ = strconv.Atoi(idStr)
		}
	} else if reObj.Match([]byte(r.URL.Path)) {
		mode = reSymb.FindString(r.URL.Path)
	}
	var param map[string]interface{} = make(map[string]interface{})
	if id > 0 {
		param["id"] = id
	}
	switch mode {
	case "users":
		u := "user"
		if formatXML {
			w.Header().Set("Content-Type", "application/xml")
			var users users
			users.Users = getFromDB(db, u, param)
			xmlB, _ := xml.MarshalIndent(users, "", "  ")
			fmt.Fprint(w, string(xmlB))
		} else {
			w.Header().Set("Content-Type", "application/json")
			jsonB, _ := json.MarshalIndent(getFromDB(db, u, param), "", "  ")
			fmt.Fprint(w, string(jsonB))
		}
	case "posts":
		p := "post"
		if formatXML {
			w.Header().Set("Content-Type", "application/xml")
			var posts posts
			posts.Posts = getFromDB(db, p, param)
			xmlB, _ := xml.MarshalIndent(posts, "", "  ")
			fmt.Fprint(w, string(xmlB))
		} else {
			w.Header().Set("Content-Type", "application/json")
			jsonB, _ := json.MarshalIndent(getFromDB(db, p, param), "", "  ")
			fmt.Fprint(w, string(jsonB))
		}
	case "comments":
		c := "comment"
		if formatXML {
			w.Header().Set("Content-Type", "application/xml")
			var comments comments
			comments.Comments = getFromDB(db, c, param)
			xmlB, _ := xml.MarshalIndent(comments, "", "  ")
			fmt.Fprint(w, string(xmlB))
		} else {
			w.Header().Set("Content-Type", "application/json")
			jsonB, _ := json.MarshalIndent(getFromDB(db, c, param), "", "  ")
			fmt.Fprint(w, string(jsonB))
		}
	default:
		body, _ := ioutil.ReadFile("./index.html")
		fmt.Fprint(w, string(body))
	}
}

func getFromDB(db *gorm.DB, obj string, param map[string]interface{}) interface{} {
	switch obj {
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

type responseStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func metodPOST(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	reObj := regexp.MustCompile(`^(\/[a-zA-Z]+)(\/)??$`) // любой /aaaaa или /aaaaa/
	//reID := regexp.MustCompile(`^(\/[a-zA-Z]+\/)\d+(\/)??$`) // любой /aaaaa/111  или /aaaaa/11111/
	reSymb := regexp.MustCompile(`[a-zA-Z]+`)
	//reNum := regexp.MustCompile(`[0-9]+`)

	if !reObj.Match([]byte(r.URL.Path)) {
		answer, _ := json.MarshalIndent(&responseStatus{Status: "error", Description: "wrong URI"}, "", "  ")
		fmt.Fprint(w, string(answer))
		return
	}
	mode := reSymb.FindString(r.URL.Path)
	switch mode {
	case "users":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		var u user
		json.Unmarshal(reqBody, &u)
		if insert2DB(db, u) == nil {
			answer, _ := json.MarshalIndent(&responseStatus{Status: "OK", Description: "OK"}, "", "  ")
			fmt.Fprint(w, string(answer))
		}
	case "posts":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		var p post
		json.Unmarshal(reqBody, &p)
		if insert2DB(db, p) == nil {
			answer, _ := json.MarshalIndent(&responseStatus{Status: "OK", Description: "OK"}, "", "  ")
			fmt.Fprint(w, string(answer))
		}
	case "comments":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		var c comment
		json.Unmarshal(reqBody, &c)
	default:
		answer, _ := json.MarshalIndent(&responseStatus{Status: "error", Description: "wrong URI"}, "", "  ")
		fmt.Fprint(w, string(answer))
	}
}

func insert2DB(db *gorm.DB, obj interface{}) error {
	objT := reflect.TypeOf(obj)
	switch strings.ToLower(objT.Name()) {
	case "user":
		var u user
		u = obj.(user)
		resU := db.Select("Name", "Username", "Email", "Phone", "Website").Create(&u)
		if resU.Error == gorm.ErrRecordNotFound {
			return resU.Error
		}
		var addr address
		addr = u.Address
		addr.UserID = u.ID
		resAddr := db.Select("UserID", "Street", "Suite", "City", "Zipcode").Create(&addr)
		if resAddr.Error == gorm.ErrRecordNotFound {
			return resAddr.Error
		}
		var comp company
		comp = u.Company
		comp.UserID = u.ID
		resComp := db.Select("UserID", "Name", "CatchPhrase", "Bs").Create(&comp)
		if resComp.Error == gorm.ErrRecordNotFound {
			return resComp.Error
		}
		var geo geo
		geo = u.Address.Geo
		geo.AddressID = addr.ID
		resGeo := db.Select("AddressID", "Lat", "Lng").Create(&geo)
		if resGeo.Error == gorm.ErrRecordNotFound {
			return resGeo.Error
		}
		return resU.Error
	case "post":
		var p post
		p = obj.(post)
		result := db.Select("UserID", "Title", "Body").Create(&p)
		return result.Error
	case "comment":
		result := db.Select("PostID", "Name", "Email", "Body").Create(obj.(comment))
		return result.Error
	}
	return nil
}

/*
{
    "Name": "gdfgdfsh",
    "Username": "dhdfghf",
    "Email": "fghfghr",
    "Address": {
      "Street": "wefwefw",
      "Suite": "wow ",
      "City": "wfwefw",
      "Zipcode": "wefwf",
      "Geo": {
        "Lat": -121.3422,
        "Lng": 23.354345
      }
    },
    "Phone": "rhrthr",
    "Website": "hrtrthr",
    "Company": {
      "Name": "dfwefw",
      "CatchPhrase": "wefwef",
      "Bs": "wefww"
    }
  }
*/
