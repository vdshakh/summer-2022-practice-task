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

const timeLayout = "15:04:05"
const maxNumberTrains = 3
const naturalNumberCondition = 0
const sortCondition = 1

type CriteriaMap map[string]struct{}

type Trains []Train

type CustomTime time.Time

// UnmarshalJSON Parses the json string in the custom format
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(timeLayout, s)
	*ct = CustomTime(nt)

	return
}

type Train struct {
	TrainID            int       `json:"trainId"`
	DepartureStationID int       `json:"departureStationId"`
	ArrivalStationID   int       `json:"arrivalStationId"`
	Price              float32   `json:"price"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	DepartureTime      time.Time `json:"departureTime"`
}

func (t *Train) UnmarshalJSON(data []byte) error {
	var aux struct {
		TrainID            int
		DepartureStationID int
		ArrivalStationID   int
		Price              float32
		ArrivalTime        CustomTime
		DepartureTime      CustomTime
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
func (t Train) String() string {
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

var validCriteria = CriteriaMap{
	"price":          {},
	"arrival-time":   {},
	"departure-time": {},
}

func main() {
	depStation, arrStation, criteria := input()

	result, err := FindTrains(depStation, arrStation, criteria)
	if err != nil {
		fmt.Printf("\nfindTrains failed: %v", err)
	}

	PrintTrains(result)
}

func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	err := validator(departureStation, arrivalStation, criteria)
	if err != nil {
		return nil, err
	}

	availableTrains, err := SelectTrains(departureStation, arrivalStation)
	if err != nil {
		return nil, fmt.Errorf("selectTrains failed: %w", err)
	}

	if len(availableTrains) < sortCondition {
		return availableTrains, nil
	}

	sortedTrains := SortTrains(availableTrains, criteria)

	if len(sortedTrains) >= maxNumberTrains {
		topTrains := SeparateTopTrains(sortedTrains)

		return topTrains, nil
	}

	return sortedTrains, nil
}

func input() (depStation, arrStation, criteria string) {
	var err error

	fmt.Print("Enter Departure Station: ")
	if depStation, err = readInput(); err != nil {
		fmt.Errorf("\nreadInput for departureStation failed: %w", err)
	}

	fmt.Print("Enter Arrival Station: ")
	if arrStation, err = readInput(); err != nil {
		fmt.Errorf("\nreadInput for arrivalStation failed: %w", err)
	}

	fmt.Print("Enter Criteria: ")
	if criteria, err = readInput(); err != nil {
		fmt.Errorf("\nreadInput for criteria failed: %w", err)
	}

	return depStation, arrStation, criteria
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	userInput, err := reader.ReadString('\n') // ReadString will block until the delimiter is entered
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)

		return "", err
	}

	userInput = strings.TrimSuffix(userInput, "\n") // remove the delimeter from the string

	return userInput, nil
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
	if len(s) == 0 {
		return fmt.Errorf("value of input is empty")
	}

	return nil
}

func validateIsNaturalNumber(s string) error {
	value, _ := strconv.Atoi(s)

	if value <= naturalNumberCondition {
		return fmt.Errorf("value is not a natural number")
	}

	return nil
}

func importInfo() (Trains, error) {
	jsonFile, err := os.Open("data.json")
	defer jsonFile.Close()

	if err != nil {
		return nil, fmt.Errorf("os.Open returns an error: %v", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var trainSchedule []Train
	if err = json.Unmarshal(byteValue, &trainSchedule); err != nil {
		return nil, fmt.Errorf("error during Unmarshal: %v", err)
	}

	return trainSchedule, nil
}

func SelectTrains(departureStation, arrivalStation string) (Trains, error) {
	var availableTrains Trains

	trainSchedule, err := importInfo()
	if err != nil {
		return nil, fmt.Errorf("importInfo failed: %w", err)
	}

	departure, _ := strconv.Atoi(departureStation)
	arrival, _ := strconv.Atoi(arrivalStation)

	for _, v := range trainSchedule {
		if v.DepartureStationID == departure && v.ArrivalStationID == arrival {
			availableTrains = append(availableTrains, v)
		}
	}

	return availableTrains, nil
}

func SortTrains(availableTrains Trains, criteria string) Trains {
	var sortedTrains Trains

	switch criteria {
	case "price":
		sortedTrains = SortTrainsByPrice(availableTrains)

	case "arrival-time":
		sortedTrains = SortTrainsByArrival(availableTrains)

	default: // aka "departure-time"
		sortedTrains = SortTrainsByDeparture(availableTrains)
	}

	return sortedTrains
}

func SortTrainsByPrice(availableTrains Trains) Trains {
	sort.SliceStable(availableTrains, func(i, j int) bool {
		if availableTrains[i].Price != availableTrains[j].Price {
			return availableTrains[i].Price < availableTrains[j].Price
		}

		return availableTrains[i].Price < availableTrains[j].Price
	})

	return availableTrains
}

func SortTrainsByArrival(availableTrains Trains) Trains {
	sort.SliceStable(availableTrains, func(i, j int) bool {
		if availableTrains[i].ArrivalTime != availableTrains[j].ArrivalTime {
			return availableTrains[i].ArrivalTime.Before(availableTrains[j].ArrivalTime)
		}

		return availableTrains[i].ArrivalTime.Before(availableTrains[j].ArrivalTime)
	})

	return availableTrains
}

func SortTrainsByDeparture(availableTrains Trains) Trains {
	sort.SliceStable(availableTrains, func(i, j int) bool {
		if availableTrains[i].DepartureTime != availableTrains[j].DepartureTime {
			return availableTrains[i].DepartureTime.Before(availableTrains[j].DepartureTime)
		}

		return availableTrains[i].DepartureTime.Before(availableTrains[j].DepartureTime)
	})

	return availableTrains
}

func SeparateTopTrains(availableTrains Trains) Trains {
	var topTrains Trains

	for i := 0; i < 3; i++ {
		topTrains = append(topTrains, availableTrains[i])
	}

	return topTrains
}

func PrintTrains(trains Trains) {
	for _, v := range trains {
		fmt.Printf("%v\n", v)
	}
}
