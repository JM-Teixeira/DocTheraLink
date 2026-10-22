package database

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/models"
	"klyniva/internal/modules/inventory"
	"klyniva/internal/modules/referral"
	"klyniva/internal/modules/schedule"
	"klyniva/internal/modules/staff"

	"gorm.io/gorm"
)

// ClimutuaClinicID is the second demo clinic: realistic mock data in PT-PT
// (staff, patients, diagnoses, referrals, schedules, sessions and stock).
const ClimutuaClinicID = "6b1f3c1e-5a8e-4c1d-9a0e-000000000002"

// ClimutuaPatients is how many patients the demo clinic has.
const ClimutuaPatients = 120

// Demo staff of Climutua (password 1234@1). Index matters: physios are
// referred to by position in the shift patterns below.
var climutuaStaff = []struct {
	username, first, last string
	role                  models.UserRole
}{
	{"cm.ana.ferreira", "Ana", "Ferreira", models.RoleCoordinator},
	{"cm.miguel.sousa", "Miguel", "Sousa", models.RoleDoctor},       // fisiatra
	{"cm.helena.carvalho", "Helena", "Carvalho", models.RoleDoctor}, // ortopedista
	{"cm.rui.teixeira", "Rui", "Teixeira", models.RoleDoctor},       // neurologista
	{"cm.ines.moreira", "Inês", "Moreira", models.RolePhysioTherapist},
	{"cm.pedro.almeida", "Pedro", "Almeida", models.RolePhysioTherapist},
	{"cm.catarina.lopes", "Catarina", "Lopes", models.RolePhysioTherapist}, // neurológica
	{"cm.joao.ribeiro", "João", "Ribeiro", models.RolePhysioTherapist},
	{"cm.mariana.costa", "Mariana", "Costa", models.RolePhysioTherapist},
	{"cm.tiago.fonseca", "Tiago", "Fonseca", models.RolePhysioTherapist},
	{"cm.sofia.martins", "Sofia", "Martins", models.RoleReceptionist},
	{"cm.carla.pinto", "Carla", "Pinto", models.RoleReceptionist},
	{"cm.teresa.gomes", "Teresa", "Gomes", models.RoleNurse},
	{"cm.rui.nogueira", "Rui", "Nogueira", models.RolePhysioTherapist}, // left the clinic (kept as a record)
}

// Skills of the Climutua physios (cm.ines.moreira … cm.tiago.fonseca).
var climutuaSkills = [][]string{
	{"Fisioterapia músculo-esquelética", "Terapia manual", "Pavimento pélvico"},
	{"Hidroterapia", "Reabilitação pós-operatória", "Fisioterapia músculo-esquelética"},
	{"Fisioterapia neurológica", "Reabilitação vestibular"},
	{"Terapia manual", "Punção seca", "Fisioterapia desportiva"},
	{"Eletroterapia", "Hidroterapia", "Drenagem linfática"},
	{"Fisioterapia desportiva", "Ligadura funcional (taping)", "Reabilitação pós-operatória"},
}

// Divisions of Climutua: name, colour, patients and credits per physio-hour.
var climutuaDivisions = []struct {
	name, color              string
	capacity, credits, floor int
}{
	{"Ginásio", "teal", 6, 8, 1},
	{"Gabinete 1", "indigo", 1, 8, 2},
	{"Gabinete 2", "amber", 1, 8, 2},
	{"Hidroterapia", "sky", 4, 8, 1},
	{"Sala de Eletroterapia", "violet", 3, 8, 2},
	{"Reabilitação Neurológica", "rose", 2, 8, 3},
}

const (
	divGinasio = iota
	divGab1
	divGab2
	divHidro
	divEletro
	divNeuro
)

// Weekly shift pattern per physio (index 0 = cm.ines.moreira); weekday 1–6.
type shiftPat struct{ weekday, start, end, div int }

var climutuaPatterns = [][]shiftPat{
	weekdays(shiftPat{0, 8, 13, divGinasio}, shiftPat{0, 14, 17, divGab1}),
	weekdays(shiftPat{0, 9, 13, divHidro}, shiftPat{0, 14, 18, divGinasio}),
	weekdays(shiftPat{0, 8, 12, divNeuro}, shiftPat{0, 13, 17, divNeuro}),
	weekdays(shiftPat{0, 10, 14, divGab2}, shiftPat{0, 15, 19, divGinasio}),
	weekdays(shiftPat{0, 8, 12, divEletro}, shiftPat{0, 13, 17, divHidro}),
	append(weekdays(shiftPat{0, 14, 20, divGinasio}), shiftPat{6, 9, 13, divGinasio}),
}

// weekdays repeats the day's shifts Monday to Friday.
func weekdays(day ...shiftPat) []shiftPat {
	var out []shiftPat
	for wd := 1; wd <= 5; wd++ {
		for _, p := range day {
			p.weekday = wd
			out = append(out, p)
		}
	}
	return out
}

// A condition the clinic treats, with what the doctor writes (PT-PT).
type condition struct {
	diagnosis, icd, reason, goals, precautions string
	doctor                                     int   // index in climutuaStaff
	physios                                    []int // physio indexes (0 = cm.ines.moreira)
	div                                        int
	careMin, careMax, perWeek, sessions        int
	minAge, maxAge                             int
}

var climutuaConditions = []condition{
	{"Lombalgia mecânica", "M54.5", "Reabilitação ativa: educação, exercício terapêutico progressivo e estabilização lombo-pélvica.",
		"Diminuir a dor; recuperar a mobilidade lombar; regresso à atividade laboral.", "Reencaminhar se surgirem sinais de alarme (anestesia em sela, alterações esfincterianas).",
		1, []int{0, 3, 5}, divGinasio, 1, 2, 2, 10, 25, 65},
	{"Cervicalgia mecânica", "M54.2", "Terapia manual, correção postural e fortalecimento dos flexores profundos do pescoço.",
		"Diminuir a dor; recuperar a amplitude cervical; reduzir as cefaleias.", "Sem manipulação cervical de alta velocidade sem rastreio prévio.",
		1, []int{0}, divGab1, 1, 1, 2, 8, 25, 60},
	{"Gonartrose bilateral", "M17.9", "Fortalecimento do quadricípite, treino de marcha e educação para a gestão da carga.",
		"Diminuir a dor; aumentar a distância de marcha; subir escadas com segurança.", "Ajustar a carga nos períodos de agudização.",
		2, []int{1, 5}, divGinasio, 2, 3, 2, 12, 60, 85},
	{"Coxartrose direita", "M16.9", "Mobilidade da anca, fortalecimento dos glúteos e treino de marcha em meio aquático.",
		"Marcha sem dor; autonomia nas atividades diárias.", "Evitar impacto repetido.",
		2, []int{1, 4}, divHidro, 2, 3, 2, 12, 60, 85},
	{"Reconstrução do ligamento cruzado anterior (pós-operatório)", "S83.5", "Reabilitação pós-operatória do LCA segundo o protocolo do cirurgião: controlo do edema, extensão completa, fortalecimento e treino neuromuscular.",
		"Extensão completa e flexão ≥ 120° à 6.ª semana; regresso ao desporto após testes funcionais.", "Cumprir as restrições do enxerto; ortótese e carga conforme indicação do cirurgião.",
		2, []int{0, 3, 5}, divGinasio, 1, 2, 3, 24, 18, 40},
	{"Capsulite adesiva do ombro esquerdo", "M75.0", "Controlo da dor e mobilização progressiva da articulação glenoumeral.",
		"Recuperar a elevação e a rotação externa; autonomia no vestir.", "Respeitar a dor na fase inflamatória.",
		1, []int{3}, divGab2, 1, 2, 2, 15, 45, 70},
	{"Síndrome da coifa dos rotadores", "M75.1", "Carga progressiva da coifa dos rotadores e controlo escapular.",
		"Atividades acima da cabeça sem dor; regresso ao trabalho.", "Vigiar dor noturna ou perda de força.",
		2, []int{0, 3, 5}, divGinasio, 1, 2, 2, 12, 35, 70},
	{"Entorse do tornozelo direito", "S93.4", "Controlo do edema, mobilidade e treino propriocetivo.",
		"Marcha normal; equilíbrio unipodal; regresso ao desporto.", "Excluir fratura (regras de Ottawa).",
		1, []int{0, 5}, divGinasio, 1, 1, 2, 8, 16, 45},
	{"Sequelas de AVC isquémico", "I69.3", "Reabilitação neurológica: controlo motor, equilíbrio, treino de transferências e de marcha; educação do cuidador.",
		"Transferências seguras; marcha com auxiliar; reduzir o risco de queda.", "Vigiar a tensão arterial; cuidados com o ombro do lado afetado; risco de queda.",
		3, []int{2}, divNeuro, 4, 5, 3, 30, 55, 88},
	{"Doença de Parkinson", "G20", "Treino de marcha com pistas externas, equilíbrio e amplitude de movimento.",
		"Manter a autonomia; reduzir os bloqueios da marcha e as quedas.", "Planear as sessões no período \"on\" da medicação.",
		3, []int{2}, divNeuro, 3, 4, 2, 20, 60, 85},
	{"Esclerose múltipla (surto recente)", "G35", "Fortalecimento, gestão da fadiga e treino de equilíbrio.",
		"Recuperar a função prévia ao surto; plano de exercício domiciliário.", "Evitar a sobrecarga térmica; gerir a fadiga.",
		3, []int{2}, divNeuro, 3, 5, 2, 20, 25, 55},
	{"Fratura do colo do fémur (pós-operatório)", "S72.0", "Reeducação da marcha, fortalecimento e treino de transferências; início em meio aquático.",
		"Marcha independente com auxiliar; prevenção de quedas.", "Carga conforme indicação do ortopedista; risco de queda.",
		2, []int{1, 4}, divHidro, 3, 4, 3, 20, 70, 92},
	{"Artroplastia total do joelho (pós-operatório)", "Z96.6", "Mobilização, fortalecimento e reeducação da marcha após prótese total do joelho.",
		"Flexão ≥ 110°; marcha sem auxiliares; subir escadas.", "Vigiar sinais de infeção ou de trombose venosa profunda.",
		2, []int{1, 5}, divGinasio, 2, 3, 3, 18, 60, 82},
	{"Epicondilite lateral do cotovelo", "M77.1", "Carga excêntrica progressiva, terapia manual e eletroterapia analgésica.",
		"Preensão sem dor; regresso ao trabalho.", "",
		1, []int{4}, divEletro, 1, 1, 2, 8, 30, 60},
	{"Fasceíte plantar", "M72.2", "Alongamentos, carga progressiva e ondas de choque.",
		"Marcha sem dor ao acordar; regresso à corrida.", "",
		1, []int{4}, divEletro, 1, 1, 2, 8, 30, 65},
	{"Hérnia discal lombar com radiculopatia", "M51.1", "Controlo da dor, mobilização neural e estabilização lombo-pélvica.",
		"Diminuir a dor irradiada; regresso às atividades diárias.", "Reencaminhar se houver défice motor progressivo.",
		1, []int{3}, divGab2, 2, 3, 2, 12, 30, 65},
	{"Síndrome do túnel cárpico", "G56.0", "Mobilização neural, ortótese noturna e exercícios de deslizamento tendinoso.",
		"Diminuir as parestesias noturnas; recuperar a força de preensão.", "",
		1, []int{4}, divEletro, 1, 1, 2, 8, 35, 70},
	{"Fibromialgia", "M79.7", "Exercício aeróbio de baixa intensidade em meio aquático e educação para a dor.",
		"Melhorar a tolerância ao esforço e o sono.", "Progressão lenta da carga.",
		1, []int{1, 4}, divHidro, 2, 2, 2, 16, 35, 65},
	{"Tendinopatia rotuliana", "M76.5", "Carga isométrica e excêntrica progressiva do tendão rotuliano.",
		"Saltar sem dor; regresso à competição.", "",
		2, []int{5, 3}, divGinasio, 1, 1, 2, 10, 18, 35},
	{"Incontinência urinária de esforço", "N39.3", "Reeducação do pavimento pélvico com biofeedback.",
		"Eliminar as perdas em esforço.", "",
		1, []int{0}, divGab1, 1, 1, 1, 10, 35, 70},
}

var (
	ptWomen    = []string{"Maria", "Ana", "Joana", "Beatriz", "Rita", "Marta", "Carolina", "Sara", "Inês", "Filipa", "Teresa", "Isabel", "Rosa", "Fernanda", "Lurdes", "Graça", "Cristina", "Patrícia", "Raquel", "Leonor"}
	ptMen      = []string{"José", "António", "Manuel", "João", "Francisco", "Luís", "Carlos", "Paulo", "Rui", "Nuno", "Ricardo", "André", "Bruno", "Diogo", "Hugo", "Vítor", "Fernando", "Joaquim", "Álvaro", "Gonçalo"}
	ptSurnames = []string{"Silva", "Santos", "Ferreira", "Pereira", "Oliveira", "Costa", "Rodrigues", "Martins", "Jesus", "Sousa", "Fernandes", "Gonçalves", "Gomes", "Lopes", "Marques", "Alves", "Almeida", "Ribeiro", "Pinto", "Carvalho", "Teixeira", "Moreira", "Correia", "Mendes", "Nunes", "Soares", "Vieira", "Monteiro", "Cardoso", "Rocha", "Neves", "Coelho", "Cruz", "Cunha", "Pires", "Ramos", "Reis", "Simões", "Antunes", "Matos"}
	ptStreets  = []string{"Rua de Cedofeita", "Rua de Santa Catarina", "Avenida da Boavista", "Rua do Almada", "Rua de Costa Cabral", "Rua da Constituição", "Avenida Fernão de Magalhães", "Rua de Pedro Hispano", "Rua Brito Capelo", "Avenida da República", "Rua Heróis de França", "Rua de Camões"}
	ptPlaces   = []struct{ postal, city string }{
		{"4050-180", "Porto"}, {"4000-443", "Porto"}, {"4100-130", "Porto"}, {"4050-024", "Porto"}, {"4200-191", "Porto"},
		{"4450-210", "Matosinhos"}, {"4430-209", "Vila Nova de Gaia"}, {"4470-157", "Maia"}, {"4460-237", "Senhora da Hora"},
	}
	ptAllergies  = []string{"Látex", "Penicilina", "Adesivos (fita de kinesio)", "AINE (ibuprofeno)"}
	ptProgress   = []string{"Boa evolução: dor 3/10, amplitude a aumentar.", "Tolerou bem o aumento de carga; sem queixas no fim da sessão.", "Sessão com mais dor (5/10) após esforço no trabalho; carga ajustada.", "Melhoria clara da marcha; já sobe escadas com corrimão.", "Plano domiciliário revisto; cumpre os exercícios diariamente."}
	ptQuestions  = []string{"Como está a tolerar o fortalecimento?", "Há sinais de melhoria na dor noturna?", "Considera necessário pedir nova ressonância?"}
	ptReplies    = []string{"Bem, dor 2/10 no fim das sessões.", "Sim, dorme a noite toda há uma semana.", "Para já não; reavalio daqui a duas semanas."}
	ptOutcomes   = []string{"Objetivos atingidos: sem dor nas atividades diárias, alta com plano de exercícios domiciliário.", "Recuperou a amplitude completa e a força; regressou ao trabalho.", "Marcha independente sem auxiliares; alta com indicação de manter atividade física regular."}
	ptDeclines   = []string{"Agenda completa nas próximas semanas; sugiro outro colega.", "Fora da minha área (necessita de fisioterapia uroginecológica)."}
	climutuaRand = rand.New(rand.NewPCG(2026, 10))
)

const climutuaName = "Climutua · Medicina Física e Reabilitação"

// seedClimutua creates the Climutua clinic once and keeps its weekly shifts
// planned four weeks ahead on every start.
func seedClimutua(now time.Time) error {
	var exists int64
	if err := DB.Model(&models.Facility{}).Where("id = ?", ClimutuaClinicID).Count(&exists).Error; err != nil {
		return err
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if exists > 0 {
		// Older databases have the name with a long dash
		if err := DB.Model(&models.Facility{}).Where("id = ?", ClimutuaClinicID).Update("name", climutuaName).Error; err != nil {
			return err
		}
		staff, divs, err := climutuaExisting()
		if err != nil || len(divs) < len(climutuaDivisions) {
			return err
		}
		_, err = planClimutuaShifts(staff, divs, today.AddDate(0, 0, -28), today.AddDate(0, 0, 28))
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		saved := DB
		DB = tx // the helpers below use DB
		defer func() { DB = saved }()
		return createClimutua(today)
	})
}

func createClimutua(today time.Time) error {
	pincode := "4050-180"
	clinic := models.Facility{ID: ClimutuaClinicID, Name: climutuaName, Code: climutuaClinicCode,
		Address: "Rua de Cedofeita 312, 4050-180 Porto", Pincode: &pincode}
	if err := DB.Create(&clinic).Error; err != nil {
		return fmt.Errorf("climutua clinic: %w", err)
	}

	staff, err := climutuaStaffIDs()
	if err != nil {
		return err
	}

	divs := make([]schedule.Division, len(climutuaDivisions))
	for i, d := range climutuaDivisions {
		divs[i] = schedule.Division{ClinicID: ClimutuaClinicID, Name: d.name, Color: d.color, Capacity: d.capacity,
			CreditLimit: d.credits, Floor: d.floor, SortOrder: i, Active: true}
		if err := DB.Create(&divs[i]).Error; err != nil {
			return fmt.Errorf("climutua division: %w", err)
		}
	}
	shifts, err := planClimutuaShifts(staff, divs, today.AddDate(0, 0, -28), today.AddDate(0, 0, 28))
	if err != nil {
		return err
	}
	items, err := createClimutuaStock(staff, today)
	if err != nil {
		return err
	}
	return createClimutuaPatients(staff, divs, shifts, items, today)
}

// climutuaStaffIDs creates the staff if needed and makes sure they work in
// Climutua; it returns their ids in climutuaStaff order.
func climutuaStaffIDs() ([]uint, error) {
	staff := make([]uint, len(climutuaStaff))
	for i, s := range climutuaStaff {
		clinicID := ClimutuaClinicID
		u := models.User{Username: s.username, Email: s.username + "@climutua.pt", FirstName: s.first, LastName: s.last,
			UserType: s.role, Password: demoPasswordHash, IsActive: true, HomeFacilityID: &clinicID}
		if err := DB.FirstOrCreate(&u, models.User{Username: s.username}).Error; err != nil {
			return nil, fmt.Errorf("climutua user %s: %w", s.username, err)
		}
		if err := DB.Model(&models.User{}).Where("id = ?", u.ID).Update("home_facility_id", clinicID).Error; err != nil {
			return nil, err
		}
		staff[i] = u.ID
	}
	for i, skills := range climutuaSkills {
		if err := linkClimutuaSkills(staff[firstPhysio+i], skills...); err != nil {
			return nil, err
		}
	}
	// cm.rui.nogueira left at the end of July; the record stays, sign-in does not
	left := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	if err := DB.Model(&models.User{}).Where("id = ? AND employment_status <> ?", staff[len(staff)-1], "left").
		Updates(map[string]any{"employment_status": "left", "employment_since": left, "is_active": false,
			"employment_note": "Mudou-se para Lisboa."}).Error; err != nil {
		return nil, err
	}
	// Start dates (a few months apart) and the history of who left
	for i, id := range staff {
		joined := time.Date(2022, 1, 10, 0, 0, 0, 0, time.UTC).AddDate(0, 2*i, 0)
		if err := DB.Model(&models.User{}).Where("id = ? AND joined_on IS NULL", id).Update("joined_on", joined).Error; err != nil {
			return nil, err
		}
	}
	if err := seedLeftEvent(staff[len(staff)-1], left, "Mudou-se para Lisboa."); err != nil {
		return nil, err
	}
	return staff, nil
}

// linkClimutuaSkills gives a Climutua physio their skills.
func linkClimutuaSkills(userID uint, names ...string) error {
	return staff.LinkSkills(DB, userID, names...)
}

// climutuaExisting reloads staff ids (by username order) and divisions.
func climutuaExisting() ([]uint, []schedule.Division, error) {
	staff, err := climutuaStaffIDs()
	if err != nil {
		return nil, nil, err
	}
	var divs []schedule.Division
	err = DB.Where("clinic_id = ?", ClimutuaClinicID).Order("sort_order").Find(&divs).Error
	return staff, divs, err
}

const firstPhysio = 4 // index of cm.ines.moreira in climutuaStaff

// planClimutuaShifts creates the weekly pattern on every date of the range
// where the physio has no shift yet. It returns all shifts of the range.
func planClimutuaShifts(staff []uint, divs []schedule.Division, from, to time.Time) ([]schedule.Shift, error) {
	var existing []schedule.Shift
	if err := DB.Where("clinic_id = ? AND date BETWEEN ? AND ?", ClimutuaClinicID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Find(&existing).Error; err != nil {
		return nil, err
	}
	has := map[string]bool{}
	for _, s := range existing {
		has[fmt.Sprintf("%s/%d", s.Date.Format("2006-01-02"), s.TherapistID)] = true
	}
	// Nobody is planned on days they are away
	var absences []schedule.Absence
	if err := DB.Where("clinic_id = ? AND \"from\" <= ? AND \"to\" >= ?", ClimutuaClinicID, to.Format("2006-01-02"), from.Format("2006-01-02")).
		Find(&absences).Error; err != nil {
		return nil, err
	}
	for _, ab := range absences {
		for d := ab.From; !d.After(ab.To); d = d.AddDate(0, 0, 1) {
			has[fmt.Sprintf("%s/%d", d.Format("2006-01-02"), ab.TherapistID)] = true
		}
	}
	out := existing
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for p, pattern := range climutuaPatterns {
			physio := staff[firstPhysio+p]
			if has[fmt.Sprintf("%s/%d", d.Format("2006-01-02"), physio)] {
				continue
			}
			for _, sp := range pattern {
				if sp.weekday != int(d.Weekday()) {
					continue
				}
				sh := schedule.Shift{ClinicID: ClimutuaClinicID, Date: d, StartHour: sp.start, EndHour: sp.end,
					TherapistID: physio, DivisionID: divs[sp.div].ID, CreatedByID: staff[0]}
				if err := DB.Create(&sh).Error; err != nil {
					return nil, fmt.Errorf("climutua shift: %w", err)
				}
				out = append(out, sh)
			}
		}
	}
	return out, nil
}

// stockSpec is one consumable of Climutua.
type stockSpec struct {
	name, category, unit, tracking string
	quantity, min, target          int
	perSession, openLife           int
}

var climutuaStock = []stockSpec{
	{"Luvas de nitrilo (M)", "EPI", "pares", inventory.TrackUnit, 1400, 150, 1200, 1, 0},
	{"Luvas de nitrilo (L)", "EPI", "pares", inventory.TrackUnit, 85, 100, 400, 0, 0},
	{"Máscaras cirúrgicas", "EPI", "unidades", inventory.TrackUnit, 260, 80, 400, 0, 0},
	{"Lençóis de marquesa descartáveis", "Higiene", "unidades", inventory.TrackUnit, 1100, 150, 1000, 1, 0},
	{"Toucas de natação", "Hidroterapia", "unidades", inventory.TrackUnit, 260, 40, 250, 0, 0},
	{"Agulhas de punção seca", "Material de tratamento", "caixas", inventory.TrackUnit, 6, 3, 12, 0, 0},
	{"Ligadura funcional (tape)", "Material de tratamento", "rolos", inventory.TrackUnit, 18, 10, 40, 0, 0},
	{"Elétrodos para TENS", "Material de tratamento", "embalagens", inventory.TrackUnit, 4, 5, 15, 0, 0},
	{"Compressas frias", "Material de tratamento", "unidades", inventory.TrackUnit, 14, 6, 20, 0, 0},
	{"Gel de ultrassons", "Material de tratamento", "frascos", inventory.TrackContainer, 7, 3, 12, 0, 30},
	{"Óleo de massagem", "Material de tratamento", "frascos", inventory.TrackContainer, 5, 2, 8, 0, 60},
	{"Solução antisséptica", "Higiene", "frascos", inventory.TrackContainer, 2, 3, 10, 0, 30},
	{"Toalhetes desinfetantes", "Higiene", "embalagens", inventory.TrackContainer, 9, 4, 20, 0, 14},
	{"Rolo de papel de marquesa", "Higiene", "rolos", inventory.TrackContainer, 12, 6, 24, 0, 0},
}

// createClimutuaStock creates the items with a little history: a delivery
// three weeks ago and containers opened by the physios.
func createClimutuaStock(staff []uint, today time.Time) (map[string]*inventory.Item, error) {
	items := map[string]*inventory.Item{}
	coord := staff[0]
	for _, s := range climutuaStock {
		it := &inventory.Item{ClinicID: ClimutuaClinicID, Name: s.name, Category: s.category, Unit: s.unit, Tracking: s.tracking,
			Quantity: s.quantity, MinQuantity: s.min, TargetQuantity: s.target, PerSession: s.perSession, OpenLifeDays: s.openLife}
		if err := DB.Omit("Rules").Create(it).Error; err != nil {
			return nil, fmt.Errorf("climutua stock %s: %w", s.name, err)
		}
		delivered := today.AddDate(0, 0, -21).Add(10 * time.Hour)
		if err := DB.Omit("User", "Item").Create(&inventory.Movement{ItemID: it.ID, Kind: inventory.KindRestock, Amount: s.quantity,
			Delta: s.quantity, QuantityAfter: s.quantity, Note: "Encomenda n.º 2026/" + fmt.Sprint(310+len(items)), UserID: coord,
			CreatedAt: delivered}).Error; err != nil {
			return nil, err
		}
		items[s.name] = it
	}
	// Containers in use, opened by physios on different days
	opened := []struct {
		item   string
		physio int
		days   int
	}{{"Gel de ultrassons", 4, 12}, {"Gel de ultrassons", 8, 40}, {"Óleo de massagem", 5, 9}, {"Solução antisséptica", 7, 5},
		{"Toalhetes desinfetantes", 10, 3}, {"Rolo de papel de marquesa", 4, 2}}
	for _, o := range opened {
		it := items[o.item]
		at := today.AddDate(0, 0, -o.days).Add(9 * time.Hour)
		it.Quantity--
		it.InUse++
		if it.OldestOpenAt == nil || at.Before(*it.OldestOpenAt) {
			it.OldestOpenAt = &at
		}
		if err := DB.Omit("User", "Item").Create(&inventory.Movement{ItemID: it.ID, Kind: inventory.KindOpen, Amount: 1, Delta: -1,
			QuantityAfter: it.Quantity, InUseAfter: it.InUse, UserID: staff[o.physio], CreatedAt: at}).Error; err != nil {
			return nil, err
		}
	}
	return items, nil
}

// slotKey identifies a physio-hour or patient-hour for capacity checks.
type slotKey struct {
	date string
	hour int
	who  string
}

// createClimutuaPatients creates patients, their referrals, threads and
// sessions (past attended/missed, future booked), respecting capacity and
// care credits, and the automatic use of gloves, sheets and swim caps.
func createClimutuaPatients(staff []uint, divs []schedule.Division, shifts []schedule.Shift,
	items map[string]*inventory.Item, today time.Time) error {
	r := climutuaRand
	// Swim caps are used per session only in the pool
	if err := DB.Create(&inventory.SessionRule{ItemID: items["Toucas de natação"].ID, DivisionID: divs[divHidro].ID, Quantity: 1}).Error; err != nil {
		return err
	}
	load := map[slotKey]int{}   // credits per physio-hour
	people := map[slotKey]int{} // patients per physio-hour
	busy := map[slotKey]bool{}  // patient-hour
	shiftFor := func(date time.Time, physio uint, div string) []int {
		var hours []int
		for _, s := range shifts {
			if s.TherapistID == physio && s.DivisionID == div && s.Date.Format("2006-01-02") == date.Format("2006-01-02") {
				for h := s.StartHour; h < s.EndHour; h++ {
					hours = append(hours, h)
				}
			}
		}
		return hours
	}

	statuses := []referral.Status{}
	// 120 patients: about 7 sessions per physio per day, like a busy clinic
	for i := 0; i < ClimutuaPatients; i++ {
		switch {
		case i < 84:
			statuses = append(statuses, referral.StatusInProgress)
		case i < 92:
			statuses = append(statuses, referral.StatusAccepted)
		case i < 107:
			statuses = append(statuses, referral.StatusCompleted)
		case i < 116:
			statuses = append(statuses, referral.StatusPending)
		default:
			statuses = append(statuses, referral.StatusDeclined)
		}
	}
	receptionist := staff[10]

	for i, status := range statuses {
		c := climutuaConditions[i%len(climutuaConditions)]
		female := r.IntN(2) == 0
		first := ptMen[r.IntN(len(ptMen))]
		gender := 1
		if female {
			first, gender = ptWomen[r.IntN(len(ptWomen))], 2
		}
		name := fmt.Sprintf("%s %s %s", first, ptSurnames[r.IntN(len(ptSurnames))], ptSurnames[r.IntN(len(ptSurnames))])
		age := c.minAge + r.IntN(c.maxAge-c.minAge+1)
		dob := time.Date(today.Year()-age, time.Month(1+r.IntN(12)), 1+r.IntN(28), 0, 0, 0, 0, time.UTC)
		year := dob.Year()
		place := ptPlaces[r.IntN(len(ptPlaces))]
		address := fmt.Sprintf("%s %d, %s %s", ptStreets[r.IntN(len(ptStreets))], 10+r.IntN(890), place.postal, place.city)
		phone := fmt.Sprintf("+351 %s%d %03d %03d", []string{"91", "92", "93", "96"}[r.IntN(4)], r.IntN(10), r.IntN(1000), r.IntN(1000))
		nif, sns := climutuaNIF(r), fmt.Sprintf("%09d", 100000000+r.IntN(899999999))
		nationality := "Portuguesa"
		clinicID := ClimutuaClinicID
		p := models.Patient{Name: name, Gender: &gender, DateOfBirth: &dob, YearOfBirth: &year, PhoneNumber: &phone,
			Address: &address, Nationality: &nationality, NIF: &nif, SNSNumber: &sns,
			CareLevel: c.careMin + r.IntN(c.careMax-c.careMin+1), FacilityID: &clinicID, IsActive: true,
			CreatedByID: &receptionist, LastEditedID: &receptionist}
		if r.IntN(6) == 0 {
			allergy := ptAllergies[r.IntN(len(ptAllergies))]
			p.Allergies = &allergy
		}
		if err := DB.Omit("Facility", "State", "District", "LocalBody", "Ward", "AssignedTo", "CreatedBy", "LastEdited").Create(&p).Error; err != nil {
			return fmt.Errorf("climutua patient: %w", err)
		}

		// The referral, written by the specialist for the condition
		doctor := staff[c.doctor]
		physio := staff[firstPhysio+c.physios[i%len(c.physios)]]
		// Completed treatments started weeks ago; open ones more recently
		daysAgo := map[referral.Status]int{referral.StatusCompleted: 35 + r.IntN(30), referral.StatusPending: r.IntN(4),
			referral.StatusAccepted: 2 + r.IntN(3), referral.StatusDeclined: 3 + r.IntN(10)}[status]
		if status == referral.StatusInProgress {
			daysAgo = 12 + r.IntN(35)
		}
		created := today.AddDate(0, 0, -daysAgo).Add(time.Duration(9+r.IntN(8)) * time.Hour)
		priority := referral.PriorityRoutine
		if c.careMax >= 4 || r.IntN(8) == 0 {
			priority = referral.PriorityUrgent
		}
		icd, reason, goals, sessions := c.icd, c.reason, c.goals, c.sessions
		ref := referral.Referral{PatientID: p.ID, FacilityID: &clinicID, DoctorID: doctor, Diagnosis: c.diagnosis, ICDCode: &icd,
			Reason: reason, Goals: &goals, RecommendedSessions: &sessions, Priority: priority, Status: status,
			CreatedAt: created, UpdatedAt: created}
		if c.precautions != "" {
			prec := c.precautions
			ref.Precautions = &prec
		}
		notes := []referral.Note{}
		note := func(author uint, kind referral.NoteKind, body string, at time.Time) {
			if at.Before(today) { // no messages from the future
				notes = append(notes, referral.Note{AuthorID: author, Kind: kind, Body: body, CreatedAt: at})
			}
		}
		accepted := created.Add(time.Duration(2+r.IntN(30)) * time.Hour)
		started := accepted.AddDate(0, 0, 1+r.IntN(3))
		var completed time.Time
		switch status {
		case referral.StatusPending:
			if i%2 == 0 { // addressed to a physio; the others are open to all
				ref.TherapistID = &physio
			}
		case referral.StatusDeclined:
			ref.TherapistID = &physio
			reason := ptDeclines[i%len(ptDeclines)]
			ref.DeclineReason = &reason
			note(physio, referral.NoteSystem, "Encaminhamento recusado: "+reason, accepted)
		default:
			ref.TherapistID = &physio
			ref.AcceptedAt = &accepted
			note(physio, referral.NoteSystem, "Encaminhamento aceite", accepted)
			if status != referral.StatusAccepted {
				ref.StartedAt = &started
				note(physio, referral.NoteSystem, "Tratamento iniciado", started)
				note(physio, referral.NoteProgress, ptProgress[r.IntN(len(ptProgress))], started.AddDate(0, 0, 7))
				if r.IntN(2) == 0 {
					q := r.IntN(len(ptQuestions))
					note(doctor, referral.NoteQuestion, ptQuestions[q], started.AddDate(0, 0, 9))
					note(physio, referral.NoteReply, ptReplies[q], started.AddDate(0, 0, 9).Add(3*time.Hour))
				}
			}
			if status == referral.StatusCompleted {
				completed = today.AddDate(0, 0, -(1 + r.IntN(4)))
				ref.CompletedAt = &completed
				outcome := ptOutcomes[r.IntN(len(ptOutcomes))]
				ref.OutcomeSummary = &outcome
				note(physio, referral.NoteSystem, "Tratamento concluído: "+outcome, completed)
			}
		}
		if err := DB.Omit("Patient", "Doctor", "Therapist").Create(&ref).Error; err != nil {
			return fmt.Errorf("climutua referral: %w", err)
		}
		for _, n := range notes {
			n.ReferralID = ref.ID
			if err := DB.Omit("Author").Create(&n).Error; err != nil {
				return err
			}
		}

		// Sessions: on the plan's weekdays, from the start of treatment
		if status != referral.StatusInProgress && status != referral.StatusAccepted && status != referral.StatusCompleted {
			continue
		}
		from := started
		end := today.AddDate(0, 0, 21)
		if status == referral.StatusAccepted {
			from = today.AddDate(0, 0, 1)
		}
		if status == referral.StatusCompleted {
			end = completed
		}
		days := map[int]bool{}
		for _, wd := range weeklyPattern(r, c.perWeek) {
			days[wd] = true
		}
		div := divs[c.div].ID
		preferred := 9 + r.IntN(8)
		for d := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC); !d.After(end); d = d.AddDate(0, 0, 1) {
			if !days[int(d.Weekday())] {
				continue
			}
			hours := shiftFor(d, physio, div)
			if len(hours) == 0 {
				continue
			}
			date := d.Format("2006-01-02")
			chosen := -1
			for k := 0; k < len(hours) && chosen < 0; k++ {
				h := hours[(k+indexOf(hours, preferred))%len(hours)]
				pk, tk := slotKey{date, h, p.ID}, slotKey{date, h, fmt.Sprint(physio)}
				if !busy[pk] && people[tk] < climutuaDivisions[c.div].capacity && load[tk]+p.CareLevel <= climutuaDivisions[c.div].credits {
					chosen = h
				}
			}
			if chosen < 0 {
				continue
			}
			tk := slotKey{date, chosen, fmt.Sprint(physio)}
			busy[slotKey{date, chosen, p.ID}] = true
			people[tk]++
			load[tk] += p.CareLevel
			ap := schedule.Appointment{ClinicID: ClimutuaClinicID, Date: d, Hour: chosen, TherapistID: physio, PatientID: p.ID,
				Credits: p.CareLevel, ReferralID: &ref.ID, Status: schedule.StatusBooked, CreatedByID: staff[0],
				CreatedAt: from.Add(-time.Hour)}
			if d.Before(today) {
				ap.Status = schedule.StatusAttended
				if r.IntN(9) == 0 {
					ap.Status = schedule.StatusMissed
				}
			}
			if err := DB.Omit("Patient").Create(&ap).Error; err != nil {
				return fmt.Errorf("climutua appointment: %w", err)
			}
			if ap.Status == schedule.StatusAttended {
				if err := useForSession(items, ap, c.div == divHidro, d.Add(time.Duration(chosen)*time.Hour+50*time.Minute)); err != nil {
					return err
				}
			}
		}
	}
	// Inês is on holiday next week; João covers and she leaves a handover
	if err := climutuaHoliday(staff, today); err != nil {
		return err
	}
	// Save the stock as it is after the sessions
	for _, it := range items {
		if err := DB.Model(&inventory.Item{}).Where("id = ?", it.ID).Updates(map[string]any{
			"quantity": it.Quantity, "in_use": it.InUse, "oldest_open_at": it.OldestOpenAt, "per_session": it.PerSession,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// climutuaHoliday records cm.ines.moreira's holiday next week through the real
// time-board rules (her patients move to cm.joao.ribeiro where he has room).
func climutuaHoliday(staff []uint, today time.Time) error {
	monday := today.AddDate(0, 0, 8-int(today.Weekday()))
	if today.Weekday() == time.Sunday {
		monday = today.AddDate(0, 0, 1)
	}
	var patients []string
	if err := DB.Model(&schedule.Appointment{}).Distinct("patient_id").
		Where("therapist_id = ? AND date BETWEEN ? AND ? AND status = ?", staff[firstPhysio], monday.Format("2006-01-02"),
			monday.AddDate(0, 0, 4).Format("2006-01-02"), schedule.StatusBooked).Limit(3).Pluck("patient_id", &patients).Error; err != nil {
		return err
	}
	notes := []string{
		"Progredir para agachamento unipodal se sem dor; manter gelo no fim.",
		"Ainda com dor noturna - privilegiar terapia manual, evitar carga acima da cabeça.",
		"Cumpre bem o plano domiciliário; reavaliar amplitude na sexta.",
	}
	cover := staff[firstPhysio+3]
	req := schedule.AbsenceRequest{TherapistID: staff[firstPhysio], CoverID: &cover, Reason: "vacation",
		From: monday.Format("2006-01-02"), To: monday.AddDate(0, 0, 4).Format("2006-01-02"), MovePatients: true,
		Handover: "Férias de uma semana. Planos de cada utente na pasta partilhada «Fisioterapia/Inês». " +
			"Os utentes do Gabinete 1 passam para o Ginásio com o João sempre que possível."}
	for i, p := range patients {
		req.Notes = append(req.Notes, schedule.HandoverInput{PatientID: p, Note: notes[i%len(notes)]})
	}
	svc := schedule.NewService(schedule.NewGormRepository(DB), clinic.NewGormResolver(DB))
	_, err := svc.CreateAbsence(context.Background(), schedule.Actor{ID: staff[0], Role: models.RoleCoordinator}, req)
	return err
}

// useForSession records what an attended session used automatically.
func useForSession(items map[string]*inventory.Item, ap schedule.Appointment, pool bool, at time.Time) error {
	use := map[string]int{"Luvas de nitrilo (M)": 1, "Lençóis de marquesa descartáveis": 1}
	if pool {
		use["Toucas de natação"] = 1
	}
	for name, n := range use {
		it := items[name]
		if it.Quantity < n {
			continue
		}
		it.Quantity -= n
		apID := ap.ID
		if err := DB.Omit("User", "Item").Create(&inventory.Movement{ItemID: it.ID, Kind: inventory.KindUse, Amount: n, Delta: -n,
			QuantityAfter: it.Quantity, Note: "session", UserID: ap.TherapistID, AppointmentID: &apID, CreatedAt: at}).Error; err != nil {
			return err
		}
	}
	return nil
}

// weeklyPattern picks the weekdays (1 = Monday) of a plan, spread like a
// real clinic's (not everyone on the same days).
func weeklyPattern(r *rand.Rand, perWeek int) []int {
	options := map[int][][]int{
		1: {{1}, {2}, {3}, {4}, {5}},
		2: {{1, 3}, {2, 4}, {3, 5}, {1, 4}, {2, 5}},
		3: {{1, 3, 5}, {2, 3, 5}, {1, 2, 4}, {1, 3, 4}},
	}[min(max(perWeek, 1), 3)]
	return options[r.IntN(len(options))]
}

func indexOf(xs []int, v int) int {
	for i, x := range xs {
		if x >= v {
			return i
		}
	}
	return 0
}

// climutuaNIF makes a valid-looking individual NIF (starts with 2, mod-11
// check digit). Mock data only.
func climutuaNIF(r *rand.Rand) string {
	digits := []int{2}
	for len(digits) < 8 {
		digits = append(digits, r.IntN(10))
	}
	sum := 0
	for i, d := range digits {
		sum += d * (9 - i)
	}
	check := 11 - sum%11
	if check >= 10 {
		check = 0
	}
	s := ""
	for _, d := range append(digits, check) {
		s += fmt.Sprint(d)
	}
	return s
}

// seedStockAccess gives the demo users their stock permissions (once per
// clinic) and puts some Climutua items on a floor:
//   - demo clinic: the nurse and Sarah change everything, Tiago nothing;
//   - Climutua: nursing everything, Inês floor 1, Mariana hygiene items on
//     floors 2 and 3, the other physios nothing (read-only).
func seedStockAccess() error {
	var users []models.User
	if err := DB.Select("id", "username").Where("username IN ?", []string{
		"de.mary.williams", "de.sarah.johnson", "cm.teresa.gomes", "cm.ines.moreira", "cm.mariana.costa",
	}).Find(&users).Error; err != nil {
		return err
	}
	id := map[string]uint{}
	for _, u := range users {
		id[u.Username] = u.ID
	}
	everything := []inventory.GrantRule{{}}
	if err := inventory.SeedGrants(DB, DemoClinicID, map[uint][]inventory.GrantRule{
		id["de.mary.williams"]: everything, id["de.sarah.johnson"]: everything,
	}); err != nil {
		return err
	}
	for name, floor := range map[string]int{"Toucas de natação": 1, "Elétrodos para TENS": 2,
		"Solução antisséptica": 2, "Toalhetes desinfetantes": 3} {
		if err := DB.Model(&inventory.Item{}).Where("clinic_id = ? AND name = ? AND floor IS NULL", ClimutuaClinicID, name).
			Update("floor", floor).Error; err != nil {
			return err
		}
	}
	return inventory.SeedGrants(DB, ClimutuaClinicID, map[uint][]inventory.GrantRule{
		id["cm.teresa.gomes"]:  everything,
		id["cm.ines.moreira"]:  {{Floors: []int{1}}},
		id["cm.mariana.costa"]: {{Floors: []int{2, 3}, Categories: []string{"Higiene"}}},
	})
}

// seedLeftEvent records in the employment history that someone left (once).
func seedLeftEvent(userID uint, since time.Time, note string) error {
	var n int64
	if err := DB.Model(&staff.EmploymentEvent{}).Where("user_id = ?", userID).Count(&n).Error; err != nil || n > 0 {
		return err
	}
	return DB.Create(&staff.EmploymentEvent{UserID: userID, Status: "left", Since: since, Note: note}).Error
}
