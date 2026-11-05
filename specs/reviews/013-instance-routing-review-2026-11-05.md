---
type: code-review
family: work
doc-type: report
reader: both
two-levels: yes
date: 2026-11-05
reviewed: specs/013-instance-routing.md
---

> Synthèse générée par le skill `two-level-report` depuis [013-instance-routing-review-2026-11-05-detailed.md](013-instance-routing-review-2026-11-05-detailed.md) : ne pas la modifier à la main.

# Revue de code · feat/instance-routing, 2026-11-05

Revue de la branche `feat/instance-routing` (spec 013) depuis `develop`. Deux tours : 33 findings au tour 1, 31 corrigés, 2 sans suite ; le tour 2 sur les corrections est sec. La question ouverte avant la revue est close.

Revue de [013-instance-routing](../013-instance-routing.md). Commande `review-loop 3 iteration`. Base `e24fad1`, HEAD de départ `ea1c44f`, HEAD de fin `70c9576`. Les SHA cités sont ceux de la branche avant regroupement des commits, conservée sur le fork : [`review/013-instance-routing-history`](https://github.com/DuxLocutus/grip-live-diff/commits/review/013-instance-routing-history). Angles : functional, tests, code-quality (par module M1 registry & pending, M2 serveur, M3 page), security, resiliency, docs. Arrêt : tour sans correction, tous les angles secs (tour 2). Contrôles : `go test -count=1 ./... && GOOS=windows go build ./... && golangci-lint run && gofmt -l . && scripts/e2e/run.sh` : ok sur `4bdc6ac` (la suite ne change que des documents de revue).

## Ce qui vous attend

Rien ne vous attend. [013-QUESTION-01](013-questions-closed.md#013-question-01) est close : option (b), appliquée après la revue.

## Findings

| Finding | Gravité | Verdict | Preuve du verdict |
|---|---|---|---|
| [W-FINDING-001](013-instance-routing-review-2026-11-05-detailed.md#w-finding-001) · La page bloque Go quand le `via` d'un lot envoyé est parti | High | corrigé `45efc5e`, cas [W-CASE-001](013-cases.md#w-case-001) | e2e, mutant sans le correctif : FAIL « batch of a gone B: Go enabled » |
| [W-FINDING-002](013-instance-routing-review-2026-11-05-detailed.md#w-finding-002) · Go accepté sans instance propriétaire : fichier ready que personne ne prend | High | corrigé `cf0e674` | test rouge sans le correctif (200 et fichier ready écrit) |
| [W-FINDING-003](013-instance-routing-review-2026-11-05-detailed.md#w-finding-003) · `readInstance` refait la lecture bornée de `loadReady` | Medium | corrigé `f3e209d` | tests inchangés, verts |
| [W-FINDING-004](013-instance-routing-review-2026-11-05-detailed.md#w-finding-004) · `pendingLines` décode le sidecar deux fois | Medium | corrigé `3e52d56` | tests inchangés, verts |
| [W-FINDING-005](013-instance-routing-review-2026-11-05-detailed.md#w-finding-005) · `tick` lit les pages ouvertes trois fois par battement | Medium | corrigé `89d9089` | tests inchangés, verts |
| [W-FINDING-006](013-instance-routing-review-2026-11-05-detailed.md#w-finding-006) · `self` déduplique les clés à la main | Medium | corrigé `858ee5b` | tests inchangés, verts |
| [W-FINDING-007](013-instance-routing-review-2026-11-05-detailed.md#w-finding-007) · `TestPendingClaim` répète le littéral `Pending{…}` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-008](013-instance-routing-review-2026-11-05-detailed.md#w-finding-008) · `pf.write` et `pf.ready` refont `writeFile` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-009](013-instance-routing-review-2026-11-05-detailed.md#w-finding-009) · La closure `pages` de `newRoutedFixtureOn` copie la map à la main | Medium | corrigé `4694622` | tests verts |
| [W-FINDING-010](013-instance-routing-review-2026-11-05-detailed.md#w-finding-010) · La règle « une instance possède un message » écrite deux fois | Medium | corrigé `bb60762` | tests inchangés, verts |
| [W-FINDING-011](013-instance-routing-review-2026-11-05-detailed.md#w-finding-011) · Chaque appelant de `replaceFile` construit le motif du temporaire | Medium | corrigé `8c3f37d` | tests verts |
| [W-FINDING-012](013-instance-routing-review-2026-11-05-detailed.md#w-finding-012) · Le harnais CDP de `manual-trigger.js` copié dans `instance-routing.js` | Medium | corrigé `fad75ac` | e2e vert |
| [W-FINDING-013](013-instance-routing-review-2026-11-05-detailed.md#w-finding-013) · La page recopie les champs de `readyState` après Go et ✕ | Medium | corrigé `3740c7d` | tests Go et e2e verts |
| [W-FINDING-014](013-instance-routing-review-2026-11-05-detailed.md#w-finding-014) · Le log de l'instance B n'est jamais affiché par `run.sh` | Medium | corrigé `31ebe67` | e2e vert |
| [W-FINDING-015](013-instance-routing-review-2026-11-05-detailed.md#w-finding-015) · L'attente « Sent → :port » écrite sept fois dans `manual-trigger.js` | Medium | corrigé `b32ff6d` | e2e vert |
| [W-FINDING-016](013-instance-routing-review-2026-11-05-detailed.md#w-finding-016) · Pas de test du flock exclusif de `--claim` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-017](013-instance-routing-review-2026-11-05-detailed.md#w-finding-017) · Pas de test du re-contrôle du trigger manuel par `--claim` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-019](013-instance-routing-review-2026-11-05-detailed.md#w-finding-019) · Pas de test du port choisi envoyé par Go | Medium | corrigé `7d77e57` | e2e, mutant sans corps du POST : FAIL |
| [W-FINDING-020](013-instance-routing-review-2026-11-05-detailed.md#w-finding-020) · Pas de test de la bannière redessinée par `poll` | Medium | corrigé `ba7d818` | e2e, mutant sans `routed`/`drawBanner` dans `poll` : FAIL |
| [W-FINDING-021](013-instance-routing-review-2026-11-05-detailed.md#w-finding-021) · Pas de test de l'absence de flèche sur un message `via` = `self` | Medium | corrigé `8ac24e9` | e2e, mutant sans le test `via === self` : FAIL |
| [W-FINDING-022](013-instance-routing-review-2026-11-05-detailed.md#w-finding-022) · Pas de test du repli d'un port mémorisé hors de `targets` | Medium | corrigé `c4007ff` | e2e, mutant sans contrôle de `targets` : FAIL |
| [W-FINDING-023](013-instance-routing-review-2026-11-05-detailed.md#w-finding-023) · « choice survives a reload » passe sans rechargement | Medium | corrigé `5b521c7` | e2e vert |
| [W-FINDING-024](013-instance-routing-review-2026-11-05-detailed.md#w-finding-024) · `contains(root, key)` : nom générique à côté de `slices.Contains` | Low | corrigé `b8a5371` | tests verts |
| [W-FINDING-025](013-instance-routing-review-2026-11-05-detailed.md#w-finding-025) · Lignes de plus de 100 colonnes dans `README.md` et `internal/pending.go` | Low | corrigé `ff00119` | plus de ligne de la branche au-delà de 100 colonnes hors bloc de code |
| [W-FINDING-026](013-instance-routing-review-2026-11-05-detailed.md#w-finding-026) · Champ de table `live2` dans `TestPeers` | Low | corrigé `933d305` | tests verts |
| [W-FINDING-027](013-instance-routing-review-2026-11-05-detailed.md#w-finding-027) · `regLive` et `liveFor` : points d'injection jamais variés | Low | corrigé `e152fbc` | tests verts |
| [W-FINDING-028](013-instance-routing-review-2026-11-05-detailed.md#w-finding-028) · `restamp` refait l'accès au dernier message | Low | corrigé `a8029e1` | tests verts |
| [W-FINDING-029](013-instance-routing-review-2026-11-05-detailed.md#w-finding-029) · `NewServer` : deux chaînes adjacentes `settingsPath, registryDir` | Low | corrigé `87be25b` | tests verts, build Windows vert |
| [W-FINDING-030](013-instance-routing-review-2026-11-05-detailed.md#w-finding-030) · `routeOf` lit `doc.live` sans accesseur, à côté de `targets()` | Low | corrigé `3a8fd76` | e2e vert |
| [W-FINDING-031](013-instance-routing-review-2026-11-05-detailed.md#w-finding-031) · En-tête du sidecar : « in manual only once claimed » ambigu | Low | corrigé `b42b448` | test de l'en-tête vert |
| [W-FINDING-032](013-instance-routing-review-2026-11-05-detailed.md#w-finding-032) · Variable `asleep` pour le cas « pas de fichier, PID vivant » | Low | fix | à corriger, lot 2 |
| [W-FINDING-018](013-instance-routing-review-2026-11-05-detailed.md#w-finding-018) | Medium | nothing | test d'un cas limite, sous High |
| [W-FINDING-033](013-instance-routing-review-2026-11-05-detailed.md#w-finding-033) | Low | nothing | hors niveau |
