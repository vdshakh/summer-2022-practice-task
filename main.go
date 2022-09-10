package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	fileName   = "data.json"
	timeLayout = "15:04:05"
)

const (
	maxNumberTrainsCondition = 3
	naturalNumberCondition   = 0
	sortCondition            = 1
)

type сriteriaMap map[string]struct{}

type trains []train

type customTime time.Time

// UnmarshalJSON Parses the json string in the custom format
func (ct *customTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(timeLayout, s)
	*ct = customTime(nt)

	return
}

type train struct {
	TrainID            int       `json:"trainId"`
	DepartureStationID int       `json:"departureStationId"`
	ArrivalStationID   int       `json:"arrivalStationId"`
	Price              float32   `json:"price"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	DepartureTime      time.Time `json:"departureTime"`
}

func (t *train) UnmarshalJSON(data []byte) error {
	var aux struct { //aux means auxiliary
		TrainID            int
		DepartureStationID int
		ArrivalStationID   int
		Price              float32
		ArrivalTime        customTime
		DepartureTime      customTime
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&aux); err != nil {
		return fmt.Errorf("decode train: %v", err)
	}

	t.TrainID = aux.TrainID
	t.DepartureStationID = aux.DepartureStationID
	t.ArrivalStationID = aux.ArrivalStationID
	t.Price = aux.Price
	t.ArrivalTime = time.Time(aux.ArrivalTime)
	t.DepartureTime = time.Time(aux.DepartureTime)

	return nil
}

// String returns the train in the custom format
func (t train) String() string {
	output := fmt.Sprintf("TrainID \t DepartureStationID \t\t ArrivalStationID \t Price \t\t\t "+
		"ArrivalTime \t\t DepartureTime \n %v\t\t %v\t\t\t\t %v\t\t\t %v\t\t\t %v\t\t %v", t.TrainID,
		t.DepartureStationID, t.ArrivalStationID, t.Price, t.ArrivalTime.Format(timeLayout),
		t.DepartureTime.Format(timeLayout))

	return output
}

var (
	criteriaErr          = errors.New("unsupported criteria")
	emptyDepartureErr    = errors.New("empty departure station")
	emptyArrivalErr      = errors.New("empty arrival station")
	badArrivalInputErr   = errors.New("bad arrival station input")
	badDepartureInputErr = errors.New("bad departure station input")
)

var validCriteria = сriteriaMap{
	"price":          {},
	"arrival-time":   {},
	"departure-time": {},
}

func main() {
	depStation := input("departureStation")
	arrStation := input("arrivalStation")
	criteria := input("criteria")

	result, err := FindTrains(depStation, arrStation, criteria)
	if err != nil {
		fmt.Printf("\nfindTrains failed: %v", err)
	}

	if len(result) > naturalNumberCondition {
		printTrains(result)
	} else {
		fmt.Printf("\ncan't find at least one train")
	}
}

func FindTrains(departureStation, arrivalStation, criteria string) (trains, error) {
	if err := validator(departureStation, arrivalStation, criteria); err != nil {
		return nil, err
	}

	availableTrains, err := selectTrains(departureStation, arrivalStation)
	if err != nil {
		return nil, fmt.Errorf("selectTrains failed: %w", err)
	}

	if len(availableTrains) < sortCondition {
		return availableTrains, nil
	}

	sortedTrains := sortTrains(availableTrains, criteria)

	if len(sortedTrains) >= maxNumberTrainsCondition {
		sortedTrains = sortedTrains[:maxNumberTrainsCondition]
	}

	return sortedTrains, nil
}

func input(parameter string) (value string) {
	var err error

	fmt.Printf("Enter %v: ", parameter)
	if value, err = readInput(); err != nil {
		fmt.Printf("\nreadInput for %v failed: %v", parameter, err)

		return ""
	}

	return value
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	userInput, err := reader.ReadString('\n') // ReadString will block until the delimiter is entered
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)

		return "", err
	}

	return strings.TrimSuffix(userInput, "\n"), nil // remove the delimeter from the string
}

func validator(departureStation, arrivalStation, criteria string) error {
	if err := validateEmpty(departureStation); err != nil {
		return emptyDepartureErr
	}

	if err := validateEmpty(arrivalStation); err != nil {
		return emptyArrivalErr
	}

	if err := validateIsNaturalNumber(departureStation); err != nil {
		return badDepartureInputErr
	}

	if err := validateIsNaturalNumber(arrivalStation); err != nil {
		return badArrivalInputErr
	}

	if _, value := validCriteria[criteria]; !value {
		return criteriaErr
	}

	return nil
}

func validateEmpty(s string) error {
	if len(s) == 0 { // if len = 0 then string is empty
		return fmt.Errorf("value of input is empty")
	}

	return nil
}

func validateIsNaturalNumber(s string) error {
	value, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("can't convert value to int: %w", err)
	}

	if value <= naturalNumberCondition {
		return fmt.Errorf("value is not a natural number")
	}

	return nil
}

func importInfo() (trains, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("os.Open returns an error: %v", err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var trainSchedule []train
	if err := json.Unmarshal(byteValue, &trainSchedule); err != nil {
		return nil, fmt.Errorf("error during Unmarshal: %v", err)
	}

	return trainSchedule, nil
}

func selectTrains(departureStation, arrivalStation string) (trains, error) {
	trainSchedule, err := importInfo()
	if err != nil {
		return nil, fmt.Errorf("importInfo failed: %w", err)
	}

	departure, err := strconv.Atoi(departureStation)
	if err != nil {
		return nil, fmt.Errorf("can't convert departureStation to int: %w", err)
	}

	arrival, err := strconv.Atoi(arrivalStation)
	if err != nil {
		return nil, fmt.Errorf("can't convert arrivalStation to int: %w", err)
	}

	var availableTrains trains

	for _, v := range trainSchedule {
		if v.DepartureStationID == departure && v.ArrivalStationID == arrival {
			availableTrains = append(availableTrains, v)
		}
	}

	return availableTrains, nil
}

func sortTrains(availableTrains trains, criteria string) trains {
	var sortedTrains trains

	switch criteria {
	case "price":
		sortedTrains = sortTrainsByPrice(availableTrains)

	case "arrival-time":
		sortedTrains = sortTrainsByArrival(availableTrains)

	case "departure-time":
		sortedTrains = sortTrainsByDeparture(availableTrains)
	}

	//default operator is useless here because we had validator func before

	return sortedTrains
}

func sortTrainsByPrice(availableTrains trains) trains {
	sort.SliceStable(availableTrains, func(i, j int) bool {
		return availableTrains[i].Price < availableTrains[j].Price
	})

	return availableTrains
}

func sortTrainsByArrival(availableTrains trains) trains {
	sort.SliceStable(availableTrains, func(i, j int) bool {
		return availableTrains[i].ArrivalTime.Before(availableTrains[j].ArrivalTime)
	})

	return availableTrains
}

func sortTrainsByDeparture(availableTrains trains) trains {
	sort.SliceStable(availableTrains, func(i, j int) bool {
		return availableTrains[i].DepartureTime.Before(availableTrains[j].DepartureTime)
	})

	return availableTrains
}

func printTrains(trains trains) {
	for _, v := range trains {
		fmt.Printf("%v\n", v)
	}
}
