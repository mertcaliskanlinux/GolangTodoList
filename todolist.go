package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
)

var db, _ = gorm.Open("mysql", "root:root@/gotodolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

//İŞLEMCİ HER ÇAĞRILDIGINDA Healtz yanıt {"alive:True"} VERECEK BİR İŞLEV YARATTIK
func Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("API Canlandırma Tamam")
	//İSTEMCİ YAZILIMININ YANITI ANLAYACAĞI ÇEKİLDE AYARLADIRK "CONTENT-TYPE,"APPLİCİTAİON/JSON"
	w.Header().Set("Content-Type", "applicitaion/json")
	io.WriteString(w, `{"alive": true}`)
}

// log GÜNLÜKÇÜ AYARLARIMIZI YAPMAK İÇİN İNİT İŞLEVİMİZİ BELİRLERDİK.
// GOLANG'DA İNİT() PROGRAM İLK BAŞLADIĞINDA YÜRÜTÜLÜR
func init() {
	//GÜNLÜK FORMATI TEXT FORMAT OLARAK AYARLADIK
	log.SetFormatter(&log.TextFormatter{})
	//SORUN BİLDİRME AÇIK
	log.SetReportCaller(true)
}

func CreateITem(w http.ResponseWriter, r *http.Request) {
	// POST İŞLEMİMDEN DEĞERİ ELDE ETTİK.
	description := r.FormValue("description")
	//GO RUN TODOLİST.GO ÇALIŞTIĞINDA ÇALIŞTIĞI TERMİNALE LOG BASICAK, DESCRİPTİON SAĞ TARAFI EKLENEN VERİYİ GÖSTERECEK
	log.WithFields(log.Fields{"description": description}).Info("Yeni Bir İtem Eklendin, Database Kaydedildi.")
	// CURL'DE VALUEYİ DESCRİPTİON OLARAK YAKALADIK
	todo := &TodoItemModel{Description: description, Completed: false}
	// TODO LİSTEYİ KALICI OLARAK OLUŞTURUP VERİ TABANINA EKLEDİK.
	db.Create(&todo)
	//son olarak, veritabanını sorgular ve işlemin başarılı olduğundan emin olmak için sorgu sonucunu istemciye döndürürüz.
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "applicitaion/json")
	json.NewEncoder(w).Encode(result.Value)
}

func UpdateITem(w http.ResponseWriter, r *http.Request) {
	// MUX'DAN BİR PARAMETRE ALIYORUZ
	// ilk başta database'de böyle bir item varmı diye kontrol ediyoruz.
	vars := mux.Vars(r)
	//strconv Paketi, SQL sorgularını çalıştırmadan önce String değişkeni bir Integer değişkene dönüştürmek  için kullanılır .
	var id, _ = strconv.Atoi(vars["id"])

	err := GetItemById(id)
	if err == false {
		w.Header().Set("Content-Type", "applicitaion/json")
		io.WriteString(w, `{"update":false, "error": "Kayıt Bulunamadı!"}`)
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Complated": completed}).Info("Kayıt Güncellendi!!")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		todo.Completed = completed
		db.Save(&todo)
		w.Header().Set("Content-Type", "applicitaion/json")
		io.WriteString(w, `{"update":true}`)

	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	// ilk başta database'de böyle bir item varmı diye kontrol ediyoruz.
	vars := mux.Vars(r)
	//strconv Paketi, SQL sorgularını çalıştırmadan önce String değişkeni bir Integer değişkene dönüştürmek  için kullanılır .
	id, _ := strconv.Atoi(vars["id"])
	err := GetItemById(id)

	if err == false {
		w.Header().Set("Content-Type", "applicitaion/json")
		io.WriteString(w, `{"deleted":false, "error": "Kayıt Bulunamadı!"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Item Silindi!")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "applicitaion/json")
		io.WriteString(w, `{deleted:true}`)
	}

}

//Bir TodoItemnesneyi güncellemek için, nesnenin gerçekten var olduğundan emin olmak için
//önce veritabanımızı sorgulayacağız. GetItemById() Bu amaçla bir fonksiyon yarattım .
func GetItemById(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoList'de Böyle Bir İtem Bulunamadı!")
		return false
	}
	return true

}

//Bir SQL SELECT sorgusu yürütecek ve istemciye geri göndermeden önce onu JSON'a kodlayacaktır.
func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Completed Items!")
	completedTodoItems := GetTodoItems(true)
	w.Header().Set("Content-Type", "applicitaion/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

//Bir SQL SELECT sorgusu yürütecek ve istemciye geri göndermeden önce onu JSON'a kodlayacaktır.
func InGetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get InCompleted Items!")
	IncompletedTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "applicitaion/json")
	json.NewEncoder(w).Encode(IncompletedTodoItems)
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("complted: ?", completed).Find(&todos).Value
	return TodoItems
}

func main() {
	defer db.Close()
	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})
	log.Info("Starting TodoList Api SERVER")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedItems).Methods("POST")
	router.HandleFunc("/todo-incompleted", InGetCompletedItems).Methods("GET")
	//yeni rotayı /todo HTTP POST isteği ile yeni CreateItem()fonksiyonumuza kaydederiz .
	router.HandleFunc("/todo", CreateITem).Methods("POST")
	router.HandleFunc("/todo{id}", UpdateITem).Methods("POST")
	router.HandleFunc("/todo{id}", DeleteItem).Methods("DELETE")

	//CORS işleyicisini mevcut uygulamamızın etrafına sarıyoruz.
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)
	http.ListenAndServe(":8000", handler)
}
