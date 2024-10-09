package main

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const CharacteristicPropertyNotify = 0x10

func main() {
	adapter := bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		fmt.Println("Erreur lors de l'activation de l'adaptateur:", err)
		return
	}

	fmt.Println("Démarrage de la découverte...")
	ch := make(chan bluetooth.ScanResult, 1)
	err = adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		fmt.Printf("Périphérique découvert: %s %d %s\n", result.Address.String(), result.RSSI, result.LocalName())
		if result.LocalName() == "HUAWEI Band 6-1AB" { // Remplacez par le nom de votre périphérique
			ch <- result
			adapter.StopScan()
		}
	})
	if err != nil {
		fmt.Println("Erreur lors du démarrage de la découverte:", err)
		return
	}

	select {
	case result := <-ch:
		device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			fmt.Println("Erreur lors de la connexion au périphérique:", err)
			return
		}
		fmt.Println("Connecté au périphérique", device.Address.String())

		// Découverte de tous les services
		services, err := device.DiscoverServices(nil)
		if err != nil {
			fmt.Println("Erreur lors de la découverte des services:", err)
			return
		}

		if len(services) == 0 {
			fmt.Println("Aucun service découvert")
		} else {
			for _, service := range services {
				fmt.Printf("Service découvert: %s\n", service.UUID().String())
				chars, err := service.DiscoverCharacteristics(nil) // Découvrir toutes les caractéristiques
				if err != nil {
					fmt.Println("Erreur lors de la découverte des caractéristiques:", err)
					continue
				}
				for _, char := range chars {
					fmt.Printf("Caractéristique découverte: %s\n", char.UUID().String())
					// Check if notifications are supported
					if char.Properties()&CharacteristicPropertyNotify != 0 {
						err = char.EnableNotifications(func(buf []byte) {
							if len(buf) > 1 {
								heartRate := buf[1]
								fmt.Printf("Données reçues: %v\n", buf)
								fmt.Printf("Fréquence cardiaque: %d bpm\n", heartRate)
							}
						})
						if err != nil {
							fmt.Println("Erreur lors de l'abonnement aux notifications:", err)
						} else {
							fmt.Println("Abonnement aux notifications réussi")
						}
					} else {
						fmt.Println("Notifications non supportées pour cette caractéristique")
					}
				}
			}
		}

		// Boucle d'attente pour maintenir le programme en cours d'exécution
		fmt.Println("En attente des notifications...")
		select {}

	case <-time.After(10 * time.Second):
		fmt.Println("Échec de la découverte du périphérique")
	}
}
