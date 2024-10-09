package main

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

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

		// Découverte des services
		services, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.NewUUID([16]byte{0x00, 0x00, 0x18, 0x0D, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB})}) // UUID du service de fréquence cardiaque
		// print les services
		fmt.Println(services)
		if err != nil {
			fmt.Println("Erreur lors de la découverte des services:", err)
			return
		}

		for _, service := range services {
			fmt.Printf("Service de fréquence cardiaque découvert: %s\n", service.UUID())
			// Découverte des caractéristiques pour le service de fréquence cardiaque
			characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{bluetooth.NewUUID([16]byte{0x00, 0x00, 0x2A, 0x37, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB})}) // UUID de la caractéristique de la fréquence cardiaque
			if err != nil {
				fmt.Println("Erreur lors de la découverte des caractéristiques:", err)
				continue
			}

			for _, characteristic := range characteristics {
				fmt.Printf("\tCaractéristique de la fréquence cardiaque découverte: %s\n", characteristic.UUID())
				// Abonnement aux notifications pour la caractéristique de la fréquence cardiaque
				err = characteristic.EnableNotifications(func(b []byte) {
					fmt.Println("Notification reçue")
					// Afficher les données brutes reçues
					fmt.Printf("Données brutes reçues: %v\n", b)
					// Vérification des données reçues
					if len(b) > 1 {
						heartRate := int(b[1])
						fmt.Printf("Fréquence cardiaque: %d bpm\n", heartRate)
					} else {
						fmt.Println("Données de notification invalides")
					}
				})
				if err != nil {
					fmt.Println("Erreur lors de l'abonnement aux notifications:", err)
					continue
				}
				fmt.Println("\tAbonnement aux notifications réussi")
			}
		}

		// Boucle d'attente pour maintenir le programme en cours d'exécution
		fmt.Println("En attente des notifications...")
		select {}

	case <-time.After(10 * time.Second):
		fmt.Println("Échec de la découverte du périphérique")
	}
}
