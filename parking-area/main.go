package main

import "fmt"

type SpaceType int

const (
	SpaceType_Compact SpaceType = iota
	SpaceType_Regular
	SpaceType_Large
)

type CarType int

const (
	CarType_Motorcycle CarType = iota
	CarType_Car
	CarType_Truck
)

// heap 實作先省略
type SpaceMinHeap struct {
	spaces []*Space
}

func (h *SpaceMinHeap) ExtractMin() *Space {

}

func (h *SpaceMinHeap) Insert(val *Space) {

}

type Vehicle struct {
	licensePlate string
	carType      CarType
}

type Space struct {
	id        int // 車位編號
	floorID   int
	spaceType SpaceType
	vehicle   *Vehicle // nil = 空位
}

func (s Space) IsEmpty() bool {
	return s.vehicle == nil
}

type Floor struct {
	id     int      // 樓層
	spaces []*Space // 全部車位
}

type ParkingLot struct {
	floor          [30]*Floor                  // 假設有 30 層樓
	carLocationMap map[string]*Space           // 車牌 → 停放位置
	availableSpots map[SpaceType]*SpaceMinHeap // 可用車位的 heap
}

func NewParkingLot() *ParkingLot {
	return &ParkingLot{
		carLocationMap: make(map[string]*Space),
		availableSpots: make(map[SpaceType]*SpaceMinHeap),
	}
}

// 分配車位
func (p *ParkingLot) Park(v *Vehicle) (*Space, error) {
	var spaceType SpaceType
	switch v.carType {
	case CarType_Motorcycle:
		spaceType = SpaceType_Compact
	case CarType_Car:
		spaceType = SpaceType_Regular
	case CarType_Truck:
		spaceType = SpaceType_Large
	default:
		return nil, fmt.Errorf("invalid car type: %v", v.carType)
	}

	spaceHeap, _ := p.availableSpots[spaceType]
	if spaceHeap == nil {
		return nil, fmt.Errorf("no available %v spot", spaceType)
	}

	space := spaceHeap.ExtractMin()
	if space == nil {
		return nil, fmt.Errorf("no available %v spot", spaceType)
	}
	space.vehicle = v
	p.carLocationMap[v.licensePlate] = space

	return space, nil
}

// 查詢位置
func (p *ParkingLot) FindCar(plate string) (*Space, error) {
	if space, found := p.carLocationMap[plate]; found {
		return space, nil
	}

	return nil, fmt.Errorf("parking lot not found")
}

// 出場釋放
func (p *ParkingLot) Leave(plate string) error {
	space, err := p.FindCar(plate)
	if err != nil {
		return err
	}

	space.vehicle = nil
	delete(p.carLocationMap, plate)
	p.availableSpots[space.spaceType].Insert(space)

	return nil
}
