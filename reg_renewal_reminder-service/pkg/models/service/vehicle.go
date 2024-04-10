package service

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/Manas8803/Puc-Detection/reg_renewal_reminder-service/pkg/models/db"
)

type Date struct {
	Year  int
	Month int
	Day   int
}
type Vehicle struct {
	OwnerName        string `json:"owner_name"`
	OfficeName       string `json:"office_name"`
	RegNo            string `json:"regno"`
	VehicleClassDesc string `json:"vehicle_class_desc"`
	Model            string `json:"model"`
	RegUpto          *Date  `json:"reg_upto"`
	VehicleType      string `json:"vehicle_type"`
	Mobile           int64  `json:"mobile"`
	PucUpto          *Date  `json:"puc_upto"`
	LastCheckDate    *Date  `json:"last_check_date"`
}

func ConvertVehicleDynToVehicle(dynVehicle db.Vehicle) (Vehicle, error) {
	reg_upto_date, err := parseDate(dynVehicle.Reg_Upto)
	if err != nil {
		return Vehicle{}, err
	}

	puc_upto_date, err := parseDate(dynVehicle.PucUpto)
	if err != nil {
		return Vehicle{}, err
	}

	mobile, err := strconv.ParseInt(dynVehicle.Mobile, 10, 64)
	if err != nil {
		return Vehicle{}, err
	}

	last_reg_date, err := parseDate(dynVehicle.LastCheckDate)
	if err != nil {
		return Vehicle{}, err
	}

	return Vehicle{
		OwnerName:        dynVehicle.OwnerName,
		OfficeName:       dynVehicle.OfficeName,
		RegNo:            dynVehicle.RegNo,
		VehicleClassDesc: dynVehicle.VehicleClassDesc,
		Model:            dynVehicle.Model,
		RegUpto:          &reg_upto_date,
		VehicleType:      dynVehicle.VehicleType,
		Mobile:           mobile,
		PucUpto:          &puc_upto_date,
		LastCheckDate:    &last_reg_date,
	}, nil
}

func IsStructEmpty(obj interface{}) bool {
	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return false
	}

	zero := reflect.Zero(value.Type())
	return value.Interface() == zero.Interface()
}

func parseDate(dateStr string) (Date, error) {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return Date{}, fmt.Errorf("invalid date format: %s", dateStr)
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return Date{}, err
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return Date{}, err
	}

	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return Date{}, err
	}

	return Date{
		Year:  year,
		Month: month,
		Day:   day,
	}, nil
}

func GetVehicleOnRegNo(reg_no string) (*Vehicle, error) {
	vehicles_db, err := db.GetVehicleOnRegNo(reg_no)
	if err != nil {
		return nil, err
	}

	if IsStructEmpty(vehicles_db) {
		return &Vehicle{}, nil
	}

	vehicle, err := ConvertVehicleDynToVehicle(*vehicles_db)
	if err != nil {
		log.Println("Error converting vehicle_dyn to vehicle")
		return nil, err
	}

	return &vehicle, nil
}