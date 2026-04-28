package demo

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/modules/rifiuti"
	"github.com/logitrack/backend/internal/repository"
)

// Demo entity ids — deterministic so the seed is idempotent and so
// the frontend can hard-link to them without depending on a
// previous list call.
const (
	DemoProduttoreID    = "demo-produttore-fro-1"
	DemoTrasportatoreID = "demo-trasportatore-fro"
	DemoDestinatarioID  = "demo-destinatario-mantova"
)

// SeedRifiuti creates the Mozzecane-flavoured demo entities for the
// rifiuti vertical: one produttore (officina industriale fittizia),
// one trasportatore modelled on FRO S.r.l., and one destinatario
// (impianto autorizzato R3 + R13 + D9). The seed is idempotent and
// safe to re-run.
func SeedRifiuti(ctx context.Context, mongo *repository.MongoRepository, tenantID string, log *zap.Logger) error {
	if tenantID == "" {
		tenantID = "demo-tenant"
	}
	now := time.Now().UTC()
	farFuture := now.Add(5 * 365 * 24 * time.Hour)

	// Produttore — fittizia officina meccanica con sede a Verona Sud.
	if _, err := mongo.GetProduttore(ctx, tenantID, DemoProduttoreID); errors.Is(err, repository.ErrNotFound) {
		p := &rifiuti.Produttore{
			ID:                 DemoProduttoreID,
			TenantID:           tenantID,
			RagioneSociale:     "Officina Meccanica Demo S.r.l.",
			CodiceFiscale:      "00000000001",
			PartitaIVA:         "00000000001",
			CodiceUnitaLocale:  "VR-DEMO-01",
			Indirizzo:          "Via dell'Industria 1",
			CAP:                "37060",
			Comune:             "Mozzecane",
			Provincia:          "VR",
			AttivitaCodiceATECO: "25.62.00",
			ContattoEmail:      "officina@demo.invalid",
		}
		if err := mongo.InsertProduttore(ctx, p); err != nil {
			log.Warn("rifiuti seed produttore failed", zap.Error(err))
		}
	} else if err != nil {
		return err
	}

	// Trasportatore — modellato su FRO S.r.l. (anagrafica reale solo
	// per riferimento, dati Albo SOSTITUITI da placeholder demo
	// finché il design partner non condivide i propri).
	if _, err := mongo.GetTrasportatore(ctx, tenantID, DemoTrasportatoreID); errors.Is(err, repository.ErrNotFound) {
		t := &rifiuti.Trasportatore{
			ID:                  DemoTrasportatoreID,
			TenantID:            tenantID,
			RagioneSociale:      "FRO S.r.l. — Demo",
			CodiceFiscale:       "03728630231",
			PartitaIVA:          "03728630231",
			AlboCategoria:       rifiuti.AlboCat5,
			AlboClasse:          rifiuti.AlboClasseE,
			AlboNumeroIscrizione: "VE/000000",
			AlboScadenza:        farFuture,
			SedeLegale:          "Via Quartieri snc, 37060 Mozzecane (VR)",
		}
		if err := mongo.InsertTrasportatore(ctx, t); err != nil {
			log.Warn("rifiuti seed trasportatore failed", zap.Error(err))
		}
	} else if err != nil {
		return err
	}

	// Destinatario — impianto fittizio in provincia di Mantova
	// autorizzato per CER demo + recupero R3 / messa in riserva R13 /
	// trattamento fisico-chimico D9.
	if _, err := mongo.GetDestinatario(ctx, tenantID, DemoDestinatarioID); errors.Is(err, repository.ErrNotFound) {
		d := &rifiuti.Destinatario{
			ID:                     DemoDestinatarioID,
			TenantID:               tenantID,
			RagioneSociale:         "Impianto Demo Recupero S.p.A.",
			CodiceFiscale:          "00000000002",
			PartitaIVA:             "00000000002",
			Indirizzo:              "Via Ecologia 10",
			Comune:                 "Goito",
			Provincia:              "MN",
			AutorizzazioneNumero:   "MN/AIA/000-DEMO",
			AutorizzazioneScadenza: farFuture,
			OperazioniAutorizzate: []rifiuti.ImpiantoOperazione{
				rifiuti.OperazioneR3,
				rifiuti.OperazioneR13,
				rifiuti.OperazioneD9,
			},
			CERAutorizzati: []rifiuti.CERCode{"150106", "130205*", "170504"},
		}
		if err := mongo.InsertDestinatario(ctx, d); err != nil {
			log.Warn("rifiuti seed destinatario failed", zap.Error(err))
		}
	} else if err != nil {
		return err
	}

	log.Info("rifiuti demo seed complete", zap.String("tenant", tenantID))
	return nil
}
