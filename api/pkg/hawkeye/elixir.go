package hawkeye

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (app *App) unlockExtraHint(w http.ResponseWriter, r *http.Request) {
	currUser := app.getUserTest(r)
	var elixir FetchedElixir
	json.NewDecoder(r.Body).Decode(&elixir)
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

	currUser := app.getUserTest(r)

	if currUser.ItemBool[fetchedElixir.Region] == false {
		app.sendResponse(w, false, Success, "A potion has already been used on this question")
		return
	}

	message, status := app.checkInventory(r, currUser, fetchedElixir)

	if !status {
		app.sendResponse(w, false, message, "You do not have this Elixir")
		return
	}

	var fetchedQuestion Question
	err := app.db.Collection("questions").FindOne(r.Context(), bson.M{"_id": fetchedElixir.QuestionID}).Decode(&fetchedQuestion)

	if err != nil {
		app.log.Errorf("Internal Server Error: %s", err.Error())
		app.sendResponse(w, false, InternalServerError, "Something went wrong")
	}

	unlockedHint := app.hangmanRemoveLetter(fetchedQuestion.Answer)

	itemBool := fmt.Sprintf("itembool.%d", fetchedQuestion.Region)

	app.db.Collection("users").FindOneAndUpdate(r.Context(), bson.M{"_id": currUser.ID},
		bson.M{
			"$set": bson.M{
				itemBool: false,
			},
		})

	app.logElixir(r, fetchedElixir, true, false)

	message, status = app.removeInventory(r, currUser, 2)

	if !status {
		app.sendResponse(w, false, message, nil)
		return
	}

	app.sendResponse(w, true, Success, unlockedHint)
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
