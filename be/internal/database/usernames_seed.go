package database

import (
	"fmt"
	"strings"

	"klyniva/internal/models"
)

// Clinic codes of the demo clinics (they prefix their people's usernames).
const (
	demoClinicCode     = "de"
	climutuaClinicCode = "cm"
)

// renamedUsers maps the demo usernames used before the naming standard
// (internal/usernames, 0.12) to the standard ones. Databases seeded by older
// versions are renamed in place, so nothing is duplicated.
var renamedUsers = map[string]string{
	"stateadmin":          "paula.neves",
	"districtadmin":       "jorge.matos",
	"facilitycoordinator": "luis.pereira",
	"doctor":              "de.john.smith",
	"physiotherapist":     "de.sarah.johnson",
	"physiotherapist2":    "de.tiago.martins",
	"nurse":               "de.mary.williams",
	"receptionist":        "de.rita.lopes",
	"coordinator":         "de.marta.ribeiro",
	"staff":               "de.nuno.dias",
	"cm.coordenacao":      "cm.ana.ferreira",
	"cm.medico1":          "cm.miguel.sousa",
	"cm.medico2":          "cm.helena.carvalho",
	"cm.medico3":          "cm.rui.teixeira",
	"cm.fisio1":           "cm.ines.moreira",
	"cm.fisio2":           "cm.pedro.almeida",
	"cm.fisio3":           "cm.catarina.lopes",
	"cm.fisio4":           "cm.joao.ribeiro",
	"cm.fisio5":           "cm.mariana.costa",
	"cm.fisio6":           "cm.tiago.fonseca",
	"cm.rececao1":         "cm.sofia.martins",
	"cm.rececao2":         "cm.carla.pinto",
	"cm.enfermagem":       "cm.teresa.gomes",
	"cm.fisio7":           "cm.rui.nogueira",
}

// applyNamingStandard gives the demo clinics their codes and renames demo
// users seeded under the old names (username and e-mail).
func applyNamingStandard() error {
	for id, code := range map[string]string{DemoClinicID: demoClinicCode, ClimutuaClinicID: climutuaClinicCode} {
		if err := DB.Model(&models.Facility{}).Where("id = ? AND code = ''", id).Update("code", code).Error; err != nil {
			return fmt.Errorf("clinic code %s: %w", code, err)
		}
	}
	for old, name := range renamedUsers {
		var taken int64
		if err := DB.Unscoped().Model(&models.User{}).Where("username = ?", name).Count(&taken).Error; err != nil {
			return err
		}
		if taken > 0 {
			continue
		}
		domain := "@klyniva.com"
		if strings.HasPrefix(name, climutuaClinicCode+".") {
			domain = "@climutua.pt"
		}
		if err := DB.Unscoped().Model(&models.User{}).Where("username = ?", old).
			Updates(map[string]any{"username": name, "email": name + domain}).Error; err != nil {
			return fmt.Errorf("rename %s: %w", old, err)
		}
	}
	return nil
}
