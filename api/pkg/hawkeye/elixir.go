package hawkeye

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//FetchedElixir ...
type FetchedElixir struct {
	//ID         primitive.ObjectID `json:"id" bson:"_id"`
	Elixir     int                `json:"elixir" bson:"elixir"`
	Region     int                `json:"region" bson:"region,omitempty"`
	ElixirName string             `bson:"elixir_name" json:"elixir_name"`
	QuestionID primitive.ObjectID `bson:"question,omitempty" json:"question"`
	QuestionNo int                `bson:"question_no" json:"question_no"`

	//Active bool               `bson:"active"  json:"active"`
}

//FetchedHint ...
type FetchedHint struct {
	ID     primitive.ObjectID `json:"id" bson:"_id"`
	Elixir int                `json:"elixir" bson:"elixir"`
	Hint   string             `json:"hint" bson:"hint"`
	Users  []string           `json:"users" bson:"users"`
}

//FetchHintRequest ...
type FetchHintRequest struct {
	Region int `json:"region" bson:"region"`
	Level  int `json:"level" bson:"level"`
}

func (app *App) getHiddenHints(w http.ResponseWriter, r *http.Request) {
	currUser := app.getUserTest(r)

	params := mux.Vars(r)

	level, err1 := strconv.Atoi(params["level"])
	region, err2 := strconv.Atoi(params["region"])

	if err1 != nil || err2 != nil {
		app.sendResponse(w, false, InternalServerError, nil)
		return
	}

	var hintRequest FetchHintRequest
	json.NewDecoder(r.Body).Decode(&hintRequest)

	var hiddenHints []Hint

	cursor, err := app.db.Collection("hiddenhints").Aggregate(r.Context(),
		bson.A{
			bson.M{"$match": bson.M{"users": currUser.ID.Hex(), "region": region, "level": level}},
		},
	)

	if err != nil {
		app.log.Errorf("Internal Server Erro: %s", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}

	if err := cursor.All(r.Context(), &hiddenHints); err != nil {
		app.log.Errorf("Internal Server Error: %s", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}

	app.sendResponse(w, true, Success, hiddenHints)
}

func (app *App) unlockExtraHint(w http.ResponseWriter, r *http.Request) {
	currUser := app.getUserTest(r)
	var elixir FetchedElixir
	json.NewDecoder(r.Body).Decode(&elixir)
	elixir.Elixir = 0
	fmt.Println(elixir)
	if currUser.ItemBool[elixir.Region] == false {
		app.sendResponse(w, false, Success, "A potion has already been used on this question")
		return
	}

	message, status := app.checkInventory(r, currUser, elixir)

	if !status {
		app.sendResponse(w, false, message, "You dont have this elixir")
		return
	}

	Filter := bson.M{"level": elixir.QuestionNo, "region": elixir.Region, "elixir": elixir.Elixir}
	var fetchedHint FetchedHint

	//fmt.Println(fetchedHint)
	if err := app.db.Collection("hiddenhints").FindOne(r.Context(), Filter).Decode(&fetchedHint); err != nil {
		app.log.Infof("ERROR %v", err.Error())
		app.sendResponse(w, false, Success, "No hidden hint yet")
		return
	}
	SetField := fmt.Sprintf("itembool.%d", elixir.Region)
	filter := bson.M{"_id": currUser.ID}
	update := bson.M{"$set": bson.M{
		SetField: false,
	},
	}
	if _, err := app.db.Collection("users").UpdateOne(r.Context(), filter, update); err != nil {
		app.log.Errorf("Database error %v", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}

	Filter = bson.M{"level": elixir.QuestionNo, "region": elixir.Region, "elixir": elixir.Elixir}
	update = bson.M{"$push": bson.M{"users": currUser.ID.Hex()}}
	if _, err := app.db.Collection("hiddenhints").UpdateOne(r.Context(), Filter, update); err != nil {
		app.log.Errorf("Database error %v", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}
	message, status = app.removeInventory(r, currUser, 0)

	if !status {
		app.sendResponse(w, false, message, "Something went wrong")
		return
	}

	app.logElixir(r, elixir, true, false)
	app.sendResponse(w, true, Success, fetchedHint)

}

func (app *App) regionMultipler(w http.ResponseWriter, r *http.Request) {
	currUser := app.getUserTest(r)

	var elixir FetchedElixir
	json.NewDecoder(r.Body).Decode(&elixir)
	elixir.Elixir = 1
	if currUser.ItemBool[elixir.Region] == false {
		app.sendResponse(w, false, Success, "A potion has already been used on this question")
		return
	}

	message, status := app.checkInventory(r, currUser, elixir)

	if !status {
		app.sendResponse(w, false, message, "You do not have this Elixir")
		return
	}

	if _, err := app.db.Collection("users").UpdateOne(
		r.Context(),
		bson.M{"_id": currUser.ID},
		bson.M{"$set": bson.M{"regionmultiplier": elixir.Region}},
	); err != nil {
		app.log.Errorf("Internal Server Error: %v", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}

	SetField := fmt.Sprintf("itembool.%d", elixir.Region)
	filter := bson.M{"_id": currUser.ID}
	update := bson.M{"$set": bson.M{
		SetField: false,
	},
	}
	if _, err := app.db.Collection("users").UpdateOne(r.Context(), filter, update); err != nil {
		app.log.Errorf("Databse error %v", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}

	message, status = app.removeInventory(r, currUser, 1)

	if !status {
		app.sendResponse(w, false, message, nil)
		return
	}

	app.log.Infof("Multiplier applied to region %d", elixir.Region)
	app.logElixir(r, elixir, true, false)
	app.sendResponse(w, true, Success, "Multiplier applied successfully")

}

func (app *App) hangMan(w http.ResponseWriter, r *http.Request) {
	var fetchedElixir FetchedElixir
	json.NewDecoder(r.Body).Decode(&fetchedElixir)
	fetchedElixir.Elixir = 2
	currUser := app.getUserTest(r)

	// Check Eligibilty
	if currUser.ItemBool[fetchedElixir.Region] == false {
		app.sendResponse(w, false, Success, "A potion has already been used on this question")
		return
	}

	// Check Inventory
	message, status := app.checkInventory(r, currUser, fetchedElixir)

	if !status {
		app.sendResponse(w, false, message, "You do not have this Elixir")
		return
	}

	// Fetch a hidden hangman hint
	Filter := bson.M{"level": fetchedElixir.QuestionNo, "region": fetchedElixir.Region, "elixir": fetchedElixir.Elixir}
	var presentHint FetchedHint
	err := app.db.Collection("hiddenhints").FindOne(r.Context(), Filter).Decode(&presentHint)

	//If no user has called the hangman hint yet, create one and push to hiddenhints
	if err == mongo.ErrNoDocuments {
		var fetchedQuestion Question
		err2 := app.db.Collection("questions").FindOne(r.Context(), bson.M{"region": fetchedElixir.Region, "level": fetchedElixir.QuestionNo}).Decode(&fetchedQuestion)

		if err2 != nil {
			app.log.Errorf("Internal Server Error: %s", err2.Error())
			app.sendResponse(w, false, InternalServerError, "Something went wrong")
		}

		unlockedHint := app.hangmanRemoveLetter(fetchedQuestion.Answer)
		newHang := Hint{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),

			Level:  fetchedElixir.QuestionNo,
			Region: fetchedElixir.Region,
			Hint:   strings.TrimSpace(unlockedHint),
			Active: false,
			Users:  []string{currUser.ID.Hex()},
		}
		_, err1 := app.db.Collection("hiddenhints").InsertOne(r.Context(), newHang)
		if err1 != nil {
			app.sendResponse(w, false, InternalServerError, "Something went wrong")
			return
		}
	}

	if err != nil && err != mongo.ErrNoDocuments {
		app.log.Errorf("Database error %v", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
		return
	}

	//Set Itembool to false, so that no other elixir can be used on this question
	itemBool := fmt.Sprintf("itembool.%d", fetchedElixir.Region)

	app.db.Collection("users").FindOneAndUpdate(r.Context(), bson.M{"_id": currUser.ID},
		bson.M{
			"$set": bson.M{
				itemBool: false,
			},
		})

	// Log the elixir
	app.logElixir(r, fetchedElixir, true, false)

	// Remove elixir from inventory
	message, status = app.removeInventory(r, currUser, 2)

	if !status {
		app.sendResponse(w, false, message, nil)
		return
	}

	app.sendResponse(w, true, Success, presentHint.Hint)
}

func (app *App) hangmanRemoveLetter(Answer string) string {

	lenAnswer := len(Answer)
	var i int
	var j int
	if lenAnswer > 5 {
		j = 5
	} else {
		j = 2
	}

	taken := make([]int, len(Answer))

	for i = 0; i < j; i++ {
		taken[i] = rand.Intn(lenAnswer)
	}

	thing := make([]string, len(Answer))
	k := 0
	for i = 0; i < lenAnswer; i++ {
		if in(taken, j, i) {
			thing[i] = string(Answer[i])
			k = k + 1
		} else {
			thing[i] = "-"
		}
	}

	hintUnlocked := strings.Join(thing, "")
	return hintUnlocked

}

func in(arr []int, n int, val int) bool {
	var i int
	for i = 0; i < n; i++ {
		if arr[i] == val {
			return true
		}
	}
	return false
}
