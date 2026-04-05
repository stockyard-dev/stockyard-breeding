package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stockyard-dev/stockyard-breeding/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}
	s.mux.HandleFunc("GET /api/animals", s.listAnimals)
	s.mux.HandleFunc("POST /api/animals", s.createAnimals)
	s.mux.HandleFunc("GET /api/animals/{id}", s.getAnimals)
	s.mux.HandleFunc("PUT /api/animals/{id}", s.updateAnimals)
	s.mux.HandleFunc("DELETE /api/animals/{id}", s.delAnimals)
	s.mux.HandleFunc("GET /api/litters", s.listLitters)
	s.mux.HandleFunc("POST /api/litters", s.createLitters)
	s.mux.HandleFunc("GET /api/litters/{id}", s.getLitters)
	s.mux.HandleFunc("PUT /api/litters/{id}", s.updateLitters)
	s.mux.HandleFunc("DELETE /api/litters/{id}", s.delLitters)
	s.mux.HandleFunc("GET /api/health_records", s.listHealthRecords)
	s.mux.HandleFunc("POST /api/health_records", s.createHealthRecords)
	s.mux.HandleFunc("GET /api/health_records/{id}", s.getHealthRecords)
	s.mux.HandleFunc("PUT /api/health_records/{id}", s.updateHealthRecords)
	s.mux.HandleFunc("DELETE /api/health_records/{id}", s.delHealthRecords)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/breeding/"})})
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listAnimals(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("species"); v != "" { filters["species"] = v }
	if v := r.URL.Query().Get("sex"); v != "" { filters["sex"] = v }
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"animals": oe(s.db.SearchAnimals(q, filters))}); return }
	wj(w, 200, map[string]any{"animals": oe(s.db.ListAnimals())})
}

func (s *Server) createAnimals(w http.ResponseWriter, r *http.Request) {
	if s.limits.MaxItems > 0 { if s.db.CountAnimals() >= s.limits.MaxItems { we(w, 402, "Free tier limit reached. Upgrade at https://stockyard.dev/breeding/"); return } }
	var e store.Animals
	json.NewDecoder(r.Body).Decode(&e)
	if e.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateAnimals(&e)
	wj(w, 201, s.db.GetAnimals(e.ID))
}

func (s *Server) getAnimals(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetAnimals(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateAnimals(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetAnimals(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Animals
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Name == "" { patch.Name = existing.Name }
	if patch.Species == "" { patch.Species = existing.Species }
	if patch.Breed == "" { patch.Breed = existing.Breed }
	if patch.Sex == "" { patch.Sex = existing.Sex }
	if patch.DateOfBirth == "" { patch.DateOfBirth = existing.DateOfBirth }
	if patch.RegistrationNumber == "" { patch.RegistrationNumber = existing.RegistrationNumber }
	if patch.SireId == "" { patch.SireId = existing.SireId }
	if patch.DamId == "" { patch.DamId = existing.DamId }
	if patch.Color == "" { patch.Color = existing.Color }
	if patch.Status == "" { patch.Status = existing.Status }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateAnimals(&patch)
	wj(w, 200, s.db.GetAnimals(patch.ID))
}

func (s *Server) delAnimals(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteAnimals(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) listLitters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"litters": oe(s.db.SearchLitters(q, filters))}); return }
	wj(w, 200, map[string]any{"litters": oe(s.db.ListLitters())})
}

func (s *Server) createLitters(w http.ResponseWriter, r *http.Request) {
	var e store.Litters
	json.NewDecoder(r.Body).Decode(&e)
	if e.SireId == "" { we(w, 400, "sire_id required"); return }
	if e.DamId == "" { we(w, 400, "dam_id required"); return }
	if e.DateBorn == "" { we(w, 400, "date_born required"); return }
	s.db.CreateLitters(&e)
	wj(w, 201, s.db.GetLitters(e.ID))
}

func (s *Server) getLitters(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetLitters(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateLitters(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetLitters(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Litters
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.SireId == "" { patch.SireId = existing.SireId }
	if patch.DamId == "" { patch.DamId = existing.DamId }
	if patch.DateBorn == "" { patch.DateBorn = existing.DateBorn }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateLitters(&patch)
	wj(w, 200, s.db.GetLitters(patch.ID))
}

func (s *Server) delLitters(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteLitters(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) listHealthRecords(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("record_type"); v != "" { filters["record_type"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"health_records": oe(s.db.SearchHealthRecords(q, filters))}); return }
	wj(w, 200, map[string]any{"health_records": oe(s.db.ListHealthRecords())})
}

func (s *Server) createHealthRecords(w http.ResponseWriter, r *http.Request) {
	var e store.HealthRecords
	json.NewDecoder(r.Body).Decode(&e)
	if e.AnimalId == "" { we(w, 400, "animal_id required"); return }
	if e.Date == "" { we(w, 400, "date required"); return }
	s.db.CreateHealthRecords(&e)
	wj(w, 201, s.db.GetHealthRecords(e.ID))
}

func (s *Server) getHealthRecords(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetHealthRecords(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateHealthRecords(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetHealthRecords(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.HealthRecords
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.AnimalId == "" { patch.AnimalId = existing.AnimalId }
	if patch.RecordType == "" { patch.RecordType = existing.RecordType }
	if patch.Date == "" { patch.Date = existing.Date }
	if patch.Description == "" { patch.Description = existing.Description }
	if patch.VetName == "" { patch.VetName = existing.VetName }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateHealthRecords(&patch)
	wj(w, 200, s.db.GetHealthRecords(patch.ID))
}

func (s *Server) delHealthRecords(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteHealthRecords(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["animals_total"] = s.db.CountAnimals()
	m["litters_total"] = s.db.CountLitters()
	m["health_records_total"] = s.db.CountHealthRecords()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "breeding"}
	m["animals"] = s.db.CountAnimals()
	m["litters"] = s.db.CountLitters()
	m["health_records"] = s.db.CountHealthRecords()
	wj(w, 200, m)
}
