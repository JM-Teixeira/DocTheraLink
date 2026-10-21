package inventory

import (
	"fmt"

	"gorm.io/gorm"
)

// SeedDemo stocks the demo clinic with typical physiotherapy consumables
// (only when it has no items yet). Some start low so the alerts show.
func SeedDemo(db *gorm.DB, clinicID string) error {
	var n int64
	if err := db.Model(&Item{}).Where("clinic_id = ?", clinicID).Count(&n).Error; err != nil || n > 0 {
		return err
	}
	const ppe, hygiene, treatment = "PPE", "Hygiene", "Treatment"
	items := []Item{
		{Name: "Surgical masks", Category: ppe, Unit: "units", Quantity: 180, MinQuantity: 50},
		{Name: "Nitrile gloves (M)", Category: ppe, Unit: "pairs", Quantity: 35, MinQuantity: 40},
		{Name: "Nitrile gloves (L)", Category: ppe, Unit: "pairs", Quantity: 90, MinQuantity: 40},
		{Name: "Hand sanitiser", Category: hygiene, Unit: "bottles", Quantity: 6, MinQuantity: 4},
		{Name: "Disinfectant wipes", Category: hygiene, Unit: "packs", Quantity: 0, MinQuantity: 3},
		{Name: "Couch roll", Category: hygiene, Unit: "rolls", Quantity: 14, MinQuantity: 6},
		{Name: "Ultrasound gel", Category: treatment, Unit: "bottles", Quantity: 3, MinQuantity: 3},
		{Name: "Massage oil", Category: treatment, Unit: "bottles", Quantity: 8, MinQuantity: 2},
		{Name: "Kinesiology tape", Category: treatment, Unit: "rolls", Quantity: 22, MinQuantity: 10},
		{Name: "TENS electrode pads", Category: treatment, Unit: "packs", Quantity: 5, MinQuantity: 4},
		{Name: "Dry needling needles", Category: treatment, Unit: "boxes", Quantity: 12, MinQuantity: 3},
		{Name: "Cold packs", Category: treatment, Unit: "units", Quantity: 10, MinQuantity: 4},
	}
	for i := range items {
		items[i].ClinicID = clinicID
		if err := db.Create(&items[i]).Error; err != nil {
			return fmt.Errorf("seed stock item %s: %w", items[i].Name, err)
		}
	}
	return nil
}
