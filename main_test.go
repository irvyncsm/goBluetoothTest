package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tinygo.org/x/bluetooth"
)

// Define the missing types
type Service interface {
	DiscoverCharacteristics(uuids []bluetooth.UUID) ([]Characteristic, error)
}

type Characteristic interface {
	EnableNotifications(callback func([]byte)) error
}

type Peripheral interface {
	DiscoverServices(uuids []bluetooth.UUID) ([]Service, error)
}

// Mock structures
type MockPeripheral struct {
	mock.Mock
}

func (m *MockPeripheral) DiscoverServices(uuids []bluetooth.UUID) ([]Service, error) {
	args := m.Called(uuids)
	return args.Get(0).([]Service), args.Error(1)
}

type MockService struct {
	mock.Mock
}

func (m *MockService) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]Characteristic, error) {
	args := m.Called(uuids)
	return args.Get(0).([]Characteristic), args.Error(1)
}

type MockCharacteristic struct {
	mock.Mock
}

func (m *MockCharacteristic) EnableNotifications(callback func([]byte)) error {
	args := m.Called(callback)
	return args.Error(0)
}

// Test function
func TestBluetoothConnection(t *testing.T) {
	mockPeripheral := new(MockPeripheral)
	mockService := new(MockService)
	mockCharacteristic := new(MockCharacteristic)

	// Mock DiscoverServices
	mockPeripheral.On("DiscoverServices", mock.Anything).Return([]Service{mockService}, nil)

	// Mock DiscoverCharacteristics
	mockService.On("DiscoverCharacteristics", mock.Anything).Return([]Characteristic{mockCharacteristic}, nil)

	// Mock EnableNotifications
	mockCharacteristic.On("EnableNotifications", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		callback := args.Get(0).(func([]byte))
		go func() {
			for i := 0; i < 5; i++ {
				callback([]byte{0, byte(60 + i)}) // Simulate heart rate data
			}
		}()
	})

	// Create a channel to receive notifications
	notifyChan := make(chan []byte)

	// Call the function to test
	go func() {
		err := connectToDevice(mockPeripheral, notifyChan)
		assert.NoError(t, err)
	}()

	// Read from the channel and print the heart rate
	for i := 0; i < 5; i++ {
		data := <-notifyChan
		heartRate := int(data[1])
		fmt.Printf("Fréquence cardiaque: %d bpm\n", heartRate)
	}

	// Assert expectations
	mockPeripheral.AssertExpectations(t)
	mockService.AssertExpectations(t)
	mockCharacteristic.AssertExpectations(t)
}

// Function to test
func connectToDevice(peripheral Peripheral, notifyChan chan []byte) error {
	services, err := peripheral.DiscoverServices([]bluetooth.UUID{bluetooth.NewUUID([16]byte{0x00, 0x00, 0x18, 0x0D, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB})})
	if err != nil {
		return fmt.Errorf("Erreur lors de la découverte des services: %w", err)
	}

	for _, service := range services {
		characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{bluetooth.NewUUID([16]byte{0x00, 0x00, 0x2A, 0x37, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB})})
		if err != nil {
			continue
		}

		for _, characteristic := range characteristics {
			err = characteristic.EnableNotifications(func(b []byte) {
				notifyChan <- b
			})
			if err != nil {
				continue
			}
		}
	}
	return nil
}
