# Toute une Ville — consignes pour Claude

Jeu idle pour Windows (Go + Ebitengine) : la ville de Clairbourg (ferme, ateliers, supermarché, école, port…).
La conceptrice est une joueuse, pas une programmeuse : réponds en français, tutoie-la, genre-toi au masculin.
Elle aime discuter des points flous avant qu'on code ; une demande claire peut être faite directement.
NOTES.md raconte l'histoire du projet et chaque décision de conception : lis-le avant de toucher à un système.

## Règles absolues

- **Ne jamais remettre la sauvegarde à zéro** (« Je le ferai moi-même si j'en ai envie ») : ne jamais augmenter
  `SaveVersion` (sim/save.go, actuellement 14). Toute nouveauté doit se charger avec une sauvegarde existante
  (champs nouveaux à zéro = valeur par défaut raisonnable ; niveaux trop hauts → bornés au chargement).
- Les commandes peuvent demander des produits qui n'existent pas encore : c'est voulu (« pour amener le joueur à
  ouvrir les nouvelles marchandises »). Ne jamais remplacer un produit demandé par un produit disponible.
- Jamais de machine derrière la caissière : décaler la deuxième rangée de machines vers la droite.
- Employés par espace = nombre de ressources exploitables (champs, machines, bâtiments d'animaux, parcs).
  Au supermarché : un magasinier par rayon.
- Une caisse sans caissière garde ses clients, qui attendent puis repartent fâchés : voulu, ne pas « corriger ».
- Tout futur lieu où des animaux mangent reçoit le point d'école « Sélection génétique » (pts_soin_rent).
- Un lieu avec beaucoup de marchandises demande plus de points de mécaniciens (camions) : voulu.
- Le jeu doit rester « aéré » : de la place entre les machines, les rayons, les zones.

## Compiler et tester

- Sous Windows (chez elle) : `go build -ldflags "-H windowsgui" -o ToutUneVille.exe .` puis lancer l'exe.
- Depuis Linux (cloud) : `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-H windowsgui -s -w" -o dist/ToutUneVille.exe .`
  (les `replace` de go.mod passent par les miroirs GitHub de golang.org/x : nécessaires dans le cloud, sans effet gênant ailleurs).
- Tests de la simulation : `go test ./sim`.
- Équilibrage : `go run ./cmd/balance N` (bot rapide) ; `go run ./cmd/balance -humain [jours]` (bot qui joue comme
  elle : séances courtes, gros gains hors ligne) ; `BALANCE_SAVE=fichier.json` écrit la sauvegarde finale.
- Captures d'écran : l'exe accepte `-captures <dossier>` (shots.go : un bot joue puis on photographie les écrans).
- Pour étudier sa partie : charger sa sauvegarde dans un test (`struct{ Jeu *Save \`json:"jeu"\` }` puis
  `g.Restore(f.Jeu, f.Jeu.Unix)`), faire tourner `g.Step(1.0/30)` et mesurer. Supprimer ces tests de débogage après.

## Organisation

- `sim/` : toute la simulation, sans dessin (lieux, IA des ouvriers, clients, camions, drones, école, commandes…).
- Racine : dessin et interface (`world.go`, `ui.go`, `*_draw.go`), sauvegarde sur disque (`save.go`), son.
- `cmd/balance` : simulateur d'équilibrage.
- Après une modification : `gofmt`, `go vet ./sim`, `go test ./sim`, compiler, ajouter une entrée à NOTES.md.
