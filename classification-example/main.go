package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/machinebox/sdk-go/classificationbox"
)

var ctx = context.Background()
var client = classificationbox.New("http://machinebox01.lxd:8080")

func prepareBox() error {
	models, err := client.ListModels(ctx)
	if err != nil {
		return err
	}
	if len(models) > 0 {
		return client.DeleteModel(ctx, models[0].ID)
	}
	return nil
}

func initModel(mID string) (classificationbox.Model, error) {
	model, err := client.CreateModel(ctx, classificationbox.NewModel(mID, "sentimentModel", "class1", "class2", "class3"))
	if err != nil {
		return model, err
	}
	return client.GetModel(ctx, model.ID)
}

func teachModel(mID string, filepath string) error {

	// todo: read training data from file ...
	buffer := &bytes.Buffer{}
	var srcFile io.ReadCloser
	var err error

	if srcFile, err = os.Open(filepath); err != nil {
		return err
	}
	defer srcFile.Close()

	if _, err = io.Copy(buffer, srcFile); err != nil {
		return err
	}

	decoder := json.NewDecoder(buffer)
	// read open bracket
	_, err = decoder.Token()
	if err != nil {
		return err
	}

	for decoder.More() {
		var example classificationbox.Example
		if err := decoder.Decode(&example); err != nil {
			return err
		}
		if err := client.Teach(ctx, mID, example); err != nil {
			return err
		}
	}
	// read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return err
	}

	return nil
}

func predict(mID, age, location, interest string) (classificationbox.PredictResponse, error) {
	request := classificationbox.PredictRequest{
		Limit: 2,
		Inputs: []classificationbox.Feature{
			{
				Key:   "user.age",
				Type:  "number",
				Value: age,
			},
			{
				Key:   "user.interests",
				Type:  "list",
				Value: interest,
			},
			{
				Key:   "user.location",
				Type:  "keyword",
				Value: location,
			},
		},
	}

	return client.Predict(ctx, mID, request)
}

func main() {
	var model classificationbox.Model
	var err error

	if err = prepareBox(); err != nil {
		fmt.Printf("error preparing box: %s\n", err)
	}

	if model, err = initModel("sentiment1"); err != nil {
		fmt.Printf("error initializing model: %s\n", err)
	}

	if err = teachModel(model.ID, "examples.json"); err != nil {
		fmt.Printf("error teaching model: %s\n", err)
	}

	if prediction, err := predict(model.ID, "56", "Aruba", "reading,coding,sailing"); err != nil {
		fmt.Printf("error in prediction: %s\n", err)
	} else {
		fmt.Printf("prediction: %#v\n", prediction)
	}
}
