package staff

import (
	"klyniva/internal/models"

	"gorm.io/gorm"
)

// SkillCatalogue is the physiotherapy skills offered on user profiles.
var SkillCatalogue = []string{
	"Fisioterapia neurológica",
	"Fisioterapia músculo-esquelética",
	"Fisioterapia desportiva",
	"Fisioterapia respiratória",
	"Fisioterapia pediátrica",
	"Pavimento pélvico",
	"Hidroterapia",
	"Terapia manual",
	"Punção seca",
	"Ligadura funcional (taping)",
	"Eletroterapia",
	"Reabilitação pós-operatória",
	"Reabilitação vestibular",
	"Drenagem linfática",
}

// SeedSkills adds the catalogue's missing skills (idempotent).
func SeedSkills(db *gorm.DB) error {
	for _, name := range SkillCatalogue {
		sk := models.Skill{Name: name}
		if err := db.Where(models.Skill{Name: name}).FirstOrCreate(&sk).Error; err != nil {
			return err
		}
	}
	return nil
}

// LinkSkills gives a user skills by name (idempotent; unknown names are skipped).
func LinkSkills(db *gorm.DB, userID uint, names ...string) error {
	for _, name := range names {
		var sk models.Skill
		if err := db.Where("name = ?", name).First(&sk).Error; err != nil {
			continue
		}
		us := models.UserSkill{UserID: userID, SkillID: sk.ID}
		if err := db.Omit("User", "Skill").Where("user_id = ? AND skill_id = ?", userID, sk.ID).FirstOrCreate(&us).Error; err != nil {
			return err
		}
	}
	return nil
}
