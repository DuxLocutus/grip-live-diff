---
type: code-review
family: work
doc-type: report
reader: both
two-levels: yes
date: 2026-11-05
reviewed: specs/013-instance-routing.md
---

<!-- synthesis -->
# Revue de code · feat/instance-routing, 2026-11-05

Revue de la branche `feat/instance-routing` (spec 013) depuis `develop`. En cours : le tour 1 donne 33 findings, 31 à corriger, 2 sans suite ; aucune question nouvelle.

Revue de [013-instance-routing](../013-instance-routing.md). Commande `review-loop 3 iteration`. Base `e24fad1`, HEAD de départ `ea1c44f`, HEAD de fin : en cours. Angles : functional, tests, code-quality (par module M1 registry & pending, M2 serveur, M3 page), security, resiliency, docs. Arrêt : en cours. Contrôles : en cours.

## Ce qui vous attend

- [013-QUESTION-01](013-questions-open.md#013-question-01) · instance vivante dont le battement échoue : possède-t-elle ses fils ? (revue informelle, R1-11)

## Findings

| Finding | Gravité | Verdict | Preuve du verdict |
|---|---|---|---|
| [W-FINDING-001](#w-finding-001) · La page bloque Go quand le `via` d'un lot envoyé est parti | High | corrigé `45efc5e`, cas [W-CASE-001](013-cases.md#w-case-001) | e2e, mutant sans le correctif : FAIL « batch of a gone B: Go enabled » |
| [W-FINDING-002](#w-finding-002) · Go accepté sans instance propriétaire : fichier ready que personne ne prend | High | corrigé `cf0e674` | test rouge sans le correctif (200 et fichier ready écrit) |
| [W-FINDING-003](#w-finding-003) · `readInstance` refait la lecture bornée de `loadReady` | Medium | corrigé `f3e209d` | tests inchangés, verts |
| [W-FINDING-004](#w-finding-004) · `pendingLines` décode le sidecar deux fois | Medium | corrigé `3e52d56` | tests inchangés, verts |
| [W-FINDING-005](#w-finding-005) · `tick` lit les pages ouvertes trois fois par battement | Medium | corrigé `89d9089` | tests inchangés, verts |
| [W-FINDING-006](#w-finding-006) · `self` déduplique les clés à la main | Medium | corrigé `858ee5b` | tests inchangés, verts |
| [W-FINDING-007](#w-finding-007) · `TestPendingClaim` répète le littéral `Pending{…}` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-008](#w-finding-008) · `pf.write` et `pf.ready` refont `writeFile` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-009](#w-finding-009) · La closure `pages` de `newRoutedFixtureOn` copie la map à la main | Medium | corrigé `4694622` | tests verts |
| [W-FINDING-010](#w-finding-010) · La règle « une instance possède un message » écrite deux fois | Medium | corrigé `bb60762` | tests inchangés, verts |
| [W-FINDING-011](#w-finding-011) · Chaque appelant de `replaceFile` construit le motif du temporaire | Medium | corrigé `8c3f37d` | tests verts |
| [W-FINDING-012](#w-finding-012) · Le harnais CDP de `manual-trigger.js` copié dans `instance-routing.js` | Medium | corrigé `fad75ac` | e2e vert |
| [W-FINDING-013](#w-finding-013) · La page recopie les champs de `readyState` après Go et ✕ | Medium | corrigé `3740c7d` | tests Go et e2e verts |
| [W-FINDING-014](#w-finding-014) · Le log de l'instance B n'est jamais affiché par `run.sh` | Medium | corrigé `31ebe67` | e2e vert |
| [W-FINDING-015](#w-finding-015) · L'attente « Sent → :port » écrite sept fois dans `manual-trigger.js` | Medium | corrigé `b32ff6d` | e2e vert |
| [W-FINDING-016](#w-finding-016) · Pas de test du flock exclusif de `--claim` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-017](#w-finding-017) · Pas de test du re-contrôle du trigger manuel par `--claim` | Medium | fix | à corriger, lot 2 |
| [W-FINDING-019](#w-finding-019) · Pas de test du port choisi envoyé par Go | Medium | corrigé `7d77e57` | e2e, mutant sans corps du POST : FAIL |
| [W-FINDING-020](#w-finding-020) · Pas de test de la bannière redessinée par `poll` | Medium | corrigé `ba7d818` | e2e, mutant sans `routed`/`drawBanner` dans `poll` : FAIL |
| [W-FINDING-021](#w-finding-021) · Pas de test de l'absence de flèche sur un message `via` = `self` | Medium | corrigé `8ac24e9` | e2e, mutant sans le test `via === self` : FAIL |
| [W-FINDING-022](#w-finding-022) · Pas de test du repli d'un port mémorisé hors de `targets` | Medium | corrigé `c4007ff` | e2e, mutant sans contrôle de `targets` : FAIL |
| [W-FINDING-023](#w-finding-023) · « choice survives a reload » passe sans rechargement | Medium | corrigé `5b521c7` | e2e vert |
| [W-FINDING-024](#w-finding-024) · `contains(root, key)` : nom générique à côté de `slices.Contains` | Low | corrigé `b8a5371` | tests verts |
| [W-FINDING-025](#w-finding-025) · Lignes de plus de 100 colonnes dans `README.md` et `internal/pending.go` | Low | corrigé `ff00119` | plus de ligne de la branche au-delà de 100 colonnes hors bloc de code |
| [W-FINDING-026](#w-finding-026) · Champ de table `live2` dans `TestPeers` | Low | corrigé `933d305` | tests verts |
| [W-FINDING-027](#w-finding-027) · `regLive` et `liveFor` : points d'injection jamais variés | Low | corrigé `e152fbc` | tests verts |
| [W-FINDING-028](#w-finding-028) · `restamp` refait l'accès au dernier message | Low | corrigé `a8029e1` | tests verts |
| [W-FINDING-029](#w-finding-029) · `NewServer` : deux chaînes adjacentes `settingsPath, registryDir` | Low | corrigé `87be25b` | tests verts, build Windows vert |
| [W-FINDING-030](#w-finding-030) · `routeOf` lit `doc.live` sans accesseur, à côté de `targets()` | Low | corrigé `3a8fd76` | e2e vert |
| [W-FINDING-031](#w-finding-031) · En-tête du sidecar : « in manual only once claimed » ambigu | Low | corrigé `b42b448` | test de l'en-tête vert |
| [W-FINDING-032](#w-finding-032) · Variable `asleep` pour le cas « pas de fichier, PID vivant » | Low | fix | à corriger, lot 2 |
| [W-FINDING-018](#w-finding-018) | Medium | nothing | test d'un cas limite, sous High |
| [W-FINDING-033](#w-finding-033) | Low | nothing | hors niveau |
<!-- /synthesis -->

## Findings en détail

<a id="w-finding-001"></a>
### W-FINDING-001 · La page bloque Go quand le `via` d'un lot envoyé est parti

- [functional] High cx:S, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:285`.
- Cas : mode manual, entrée ouverte routée vers B, Go vers B (ready avec `via` = B, `pid` null), puis B s'arrête. Dans la page de A, `ready.state` = "sent" : `queued()` est vrai, `go.disabled = queued() || !open` vaut true, le bouton affiche « Sent → :B » sans « (gone) ». Rien n'écrit le sidecar ni le ready file : aucun Monitor ne se réveille. Le spec 013 dit : « if its `via` is gone, it rewrites `via` to that port, so the Monitors wake » (AC11). La page ne peut pas déclencher cette réécriture.
- Preuve : binaire de HEAD, HOME et XDG_CACHE_HOME temporaires, A :18531 et B :18532 sur le même `docs/` ; PUT vers B, passage en manual, POST ready vers B, `kill -TERM` B, GET sur A → `ready {'state': 'sent', 'pid': None, ..., 'via': 18532}`, `live [18531]`. Go désactivé : lecture de `annotate.js:263-265` et `:285` (non rejoué dans Chrome).
- Correctif proposé (inst) : Go actif quand `ready.state === "sent"` et `ready.via` absent de `doc.live` ; libellé « Sent → :<via> (gone) » comme `routeOf`. Class (cx:M) : une étape de `scripts/e2e/instance-routing.js`, Go vers B puis arrêt de B, qui vérifie Go actif et le ready file à `via` = A.
- Verdict : corrigé.
- Fix : Go actif sur un lot `sent` dont le `via` n'est pas dans `live`, libellé « Sent → :<via> (gone) » ; étape e2e : Go vers B, arrêt de B, Go actif, ready file réécrit à `via` = A. `inst`, cas [W-CASE-001](013-cases.md#w-case-001). Commits `45efc5e`.

<a id="w-finding-002"></a>
### W-FINDING-002 · Go accepté sans instance propriétaire : fichier ready que personne ne prend

- [functional] High cx:S, origine `fix` (`f7b1b60`), `internal/annotations.go:388`.
- Cas : A lancée sur `docs/specs/`, page `docs/other.md` servie par `../` (cas de AC8), aucune autre instance, trigger manuel, une entrée ouverte. `pick` rend 0, `owner(via)` rend 0, `writeRoute.owns` compare 0 à 0 et rend vrai. Go → 200 « Sent », fichier ready sans `via`. L'owner rule ne l'attribue à aucune instance : `pending` ne le liste pour personne, la page affiche « Sent » sans fin. Le spec dit : « Go is refused: when no open entry is owned by that port ».
- Preuve : test jetable dans un worktree (`sendReady` avec `peers{self:6419, live:[{Port:6419, Root:"/elsewhere"}]}`, `rt` construit comme `Server.route`) → `via=0 st={State:sent … Via:0} err=<nil> ready="{\"sent_at\":\"2026-11-05T07:51:50+01:00\",\"pid\":null}\n"`.
- Gravité corrigée : Medium → High (doctrine, High : le code diffère d'une exigence, sur un cas que AC8 nomme).
- Correctif proposé (class) : `owns` rend faux quand le routage est actif et `rt.via == 0` (`rt.owner == nil || (rt.via != 0 && rt.owner(msgVia) == rt.via)`) : Go refusé en 409. Un cas de test de `sendReady` : racine qui ne contient pas le fichier → refus, pas de fichier ready.
- Verdict : corrigé.
- Fix : `writeRoute.owns` rend faux quand `via` = 0 ; test `TestGoWithNoOwnerIsRefused` (409, pas de fichier ready). `class`. Commits `cf0e674`.

<a id="w-finding-003"></a>
### W-FINDING-003 · `readInstance` refait la lecture bornée de `loadReady`

- [code-quality] Medium cx:S, origine `new` (`9845bb3`), `internal/registry.go:101`.
- Cas : `readInstance` refait `os.Open`, `io.ReadAll(io.LimitReader(f, max))`, close, `Unmarshal` de `loadReady` (`internal/ready.go:104`) ; la fermeture diverge déjà (`//nolint:errcheck defer f.Close()` contre `_ = f.Close()`). La prochaine règle de lecture d'un fichier tiers s'écrit deux fois.
- Preuve : `grep -n "io.LimitReader" internal/*.go` → `ready.go:110`, `registry.go:107`.
- Correctif proposé (class) : un helper `readLimited(p string, max int64) ([]byte, error)`, utilisé par `readInstance` et `loadReady` ; chaque appelant garde son décodeur.
- Verdict : corrigé.
- Fix : helper `readLimited`, utilisé par `readInstance` et `loadReady`. `class`. Commits `f3e209d`.

<a id="w-finding-004"></a>
### W-FINDING-004 · `pendingLines` décode le sidecar deux fois

- [code-quality] Medium cx:M, origine `new` (`2b67210`), `internal/pending.go:145`.
- Cas : `pendingLines` décode le sidecar une fois dans `head` pour `file`, puis par `parseSidecar`, qui écrase `sc.File` avec `path.Base(target)` (`internal/annotations.go:313`). Un sidecar de 1 MiB est décodé deux fois, sous le flock exclusif avec `--claim`.
- Preuve : lecture de `internal/pending.go:150` et `internal/annotations.go:313`.
- Gravité corrigée : Low → Medium (doctrine, Medium : solution plus simple aux mêmes critères).
- Correctif proposé (inst) : `parseSidecar` rend aussi le nom stocké avant l'écrasement ; `pendingLines` le lit sans second décodage.
- Verdict : corrigé.
- Fix : `parseSidecar` rend le nom stocké ; `pendingLines` ne décode plus deux fois. `inst`. Commits `3e52d56`.

<a id="w-finding-005"></a>
### W-FINDING-005 · `tick` lit les pages ouvertes trois fois par battement

- [code-quality] Medium cx:S, origine `new` (`9845bb3`), `internal/registry.go:336`.
- Cas : `tick` appelle `g.pages()` par `write` → `self`, par `read` → `self`, puis `open := g.pages()` (`:353`). Une page ouverte entre deux appels est dans `open` et pas dans `self().Pages`. Sans effet aujourd'hui (`targets` ne lit que `Root` pour soi), mais le lecteur doit le prouver.
- Preuve : lecture de `internal/registry.go:283`, `:316`, `:336`, `:353`.
- Gravité corrigée : Low → Medium (doctrine, Medium : défaut de structure qui causera un bug ; solution plus simple).
- Correctif proposé (inst) : un seul instantané `open` dans `tick`, `self` construit depuis lui et passé à `write` et `read`.
- Verdict : corrigé.
- Fix : un seul instantané des pages dans `tick`, passé à `write` et `readWith`. `inst`. Commits `89d9089`.

<a id="w-finding-006"></a>
### W-FINDING-006 · `self` déduplique les clés à la main

- [code-quality] Medium cx:S, origine `new` (`9845bb3`), `internal/registry.go:283`.
- Cas : `self` déduplique avec `slices.Contains` dans une boucle, puis trie : O(n²) et cinq lignes pour `slices.Compact(slices.Sorted(maps.Values(…)))`.
- Preuve : lecture de `internal/registry.go:287-295`.
- Gravité corrigée : Low → Medium (doctrine, Medium : implémentation parallèle d'un utilitaire existant).
- Correctif proposé (inst) : `keys := slices.Compact(slices.Sorted(maps.Values(g.pages())))`, `[]string{}` si nil pour garder `[]` en JSON.
- Verdict : corrigé.
- Fix : `slices.Compact(slices.Sorted(maps.Values(open)))`, `[]` si vide. `inst`. Commits `858ee5b`.

<a id="w-finding-007"></a>
### W-FINDING-007 · `TestPendingClaim` répète le littéral `Pending{…}`

- [code-quality] Medium cx:S, origine `new` (`2b67210`), `internal/pending_test.go:192`.
- Cas : le littéral complet `Pending{Port: portA, PID: pf.pidA, Claim: agent, Registry: pf.reg, Paths: …, Out: &out, Err: &errOut}` est écrit quatre fois, plus `:155`, à côté de `pf.run`. Un nouveau champ de `Pending` change cinq endroits.
- Preuve : `internal/pending_test.go:202`, `:211`, `:224`, `:227`, `:155`.
- Gravité corrigée : Low → Medium (doctrine, Medium : logique dupliquée).
- Correctif proposé (inst) : une méthode `pf.pending(port, pid, claim, paths...)` utilisée par `run` et par ces appels.
- Verdict : fix.
- Fix : méthode `pf.pending` qui construit `Pending` et ses tampons, utilisée par `run` et `TestPendingClaim`, `TestPendingResolvesThePath`. `inst`. Commits `1952c83`.

<a id="w-finding-008"></a>
### W-FINDING-008 · `pf.write` et `pf.ready` refont `writeFile`

- [code-quality] Medium cx:S, origine `new` (`2b67210`), `internal/pending_test.go:35`.
- Cas : `pf.write` et `pf.ready` refont `os.WriteFile` + `t.Fatal`, alors que le même fichier utilise le helper `writeFile(t, dir, name, content)` (`internal/reload_test.go:371`).
- Preuve : `internal/pending_test.go:51`, `:57` contre `:139`, `:238`, `:257`.
- Gravité corrigée : Low → Medium (doctrine, Medium : implémentation parallèle d'un utilitaire existant).
- Correctif proposé (inst) : `writeFile(pf.t, pf.dir, …)` dans `write` et `ready`.
- Verdict : fix.
- Fix : `pf.write` et `pf.ready` passent par `writeFile`. `inst`. Commits `1952c83`.

<a id="w-finding-009"></a>
### W-FINDING-009 · La closure `pages` de `newRoutedFixtureOn` copie la map à la main

- [code-quality] Medium cx:S, origine `new` (`9845bb3`), `internal/registry_test.go:568`.
- Cas : cinq lignes de copie là où `maps.Clone(rf.open)` suffit.
- Preuve : lecture de `internal/registry_test.go:568-576`.
- Gravité corrigée : Low → Medium (doctrine, Medium : implémentation parallèle d'un utilitaire existant).
- Correctif proposé (inst) : `return maps.Clone(rf.open)` sous le verrou.
- Verdict : corrigé.
- Fix : `maps.Clone(rf.open)`. `inst`. Commits `4694622`.

<a id="w-finding-010"></a>
### W-FINDING-010 · La règle « une instance possède un message » écrite deux fois

- [code-quality] Medium cx:S, origine `fix` (`f7b1b60`), `internal/annotations.go:387`.
- Cas : `writeRoute.owns` (`rt.owner == nil || rt.owner(msgVia) == rt.via`) et `Pending.owns` (`internal/pending.go:72` « return func(string, int) bool { return true } », `:83` « v.owner(key, via) == p.Port ») encodent chacun la règle et « routage coupé possède tout ». Une évolution faite d'un côté (par exemple [W-FINDING-002](#w-finding-002)) laisse `pending` lister ce que le serveur refuse, ou l'inverse.
- Preuve : lecture de `internal/annotations.go:381-398`, `internal/pending.go:67-84`.
- Correctif proposé (class) : une méthode `peers.owns(key string, via, self int) bool` dans `registry.go`, appelée par `writeRoute` et `Pending.owns`.
- Verdict : corrigé.
- Fix : `peers.owns(key, via, port)` appelé par `writeRoute.owns` et `Pending.owns` ; `writeRoute` porte la lecture du registre. `class`. Commits `bb60762`.

<a id="w-finding-011"></a>
### W-FINDING-011 · Chaque appelant de `replaceFile` construit le motif du temporaire

- [code-quality] Medium cx:S, origine `new` (`2b67210`), `internal/ready.go:124`.
- Cas : `writeReady` et `registry.write` construisent `"."+filepath.Base(p)+"-*"` ; le point initial est une contrainte (`internal/registry.go:33` « a temp file of replaceFile starts with a dot »). Un nouvel appelant sans point montre son temporaire dans le répertoire servi, et ce commentaire devient faux sans signal.
- Preuve : `grep -n replaceFile internal/*.go` → `ready.go:124`, `registry.go:316` (même motif), `settings.go:87` (`.settings-*.json`).
- Correctif proposé (class) : `replaceFile(path, mode, data)` dérive le motif ; supprimer le paramètre `pattern` aux trois appels.
- Verdict : corrigé.
- Fix : `replaceFile(path, mode, data)` dérive le motif `.<base>-*` ; le temporaire des réglages devient `.settings.json-*` (glob du test aligné). `class`. Commits `8c3f37d`.

<a id="w-finding-012"></a>
### W-FINDING-012 · Le harnais CDP de `manual-trigger.js` copié dans `instance-routing.js`

- [code-quality] Medium cx:M, origine `new` (`715401f`), `scripts/e2e/instance-routing.js:18`.
- Cas : environ 70 lignes copiées (send, ev, sleep, waitFor, check, garde beforeExit, onmessage/onclose), et la copie diverge : sans gestionnaire `Page.javascriptDialogOpening`, une boîte de dialogue dans le run 013 donne un timeout de 10 s par appel CDP au lieu de « FAIL no browser dialog » ; trois commentaires de contrainte perdus ; pas de `waitFor` à 200 ms au lieu de 100 ms.
- Preuve : `diff <(sed -n 21,72p manual-trigger.js) <(sed -n 18,62p instance-routing.js)` → seules différences : les 3 commentaires, `sleep(100)`/`sleep(200)` ; `grep -c javascriptDialogOpening` → manual-trigger.js:1, instance-routing.js:0.
- Correctif proposé (class) : extraire `scripts/e2e/cdp.js` (connect, send, ev, waitFor, check, beforeExit, dialog, erreurs console), chargé par les deux scripts avec `require`.
- Verdict : corrigé.
- Fix : `scripts/e2e/cdp.js` (connect, send, ev, waitFor, check, beforeExit, dialogue, erreurs console), chargé par les deux scripts. `class`. Commits `fad75ac`.

<a id="w-finding-013"></a>
### W-FINDING-013 · La page recopie les champs de `readyState` après Go et ✕

- [code-quality] Medium cx:M, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:393`.
- Cas : `doc.ready = { state: st.state, pid: st.pid, sent_at: st.sent_at, via: st.via }` recopie les champs de `readyState`, parce que `readyResponse` (`internal/ready.go:64`) aplatit `readyState` avec `routingInfo` alors que le GET le range sous `ready`. Un champ ajouté à `readyState` disparaît après Go ou ✕ jusqu'au fetch suivant. Le commentaire de `readyResponse` « as the annotations GET has it » dit le contraire de la forme réelle.
- Preuve : `grep -n "doc.ready = " annotate.js` → 335 `doc.ready = d.ready;`, 393 la reconstruction champ par champ ; avant la branche (`e24fad1`), `doc.ready = st`.
- Correctif proposé (class) : un champ `Ready readyState` (clé JSON `ready`) dans `readyResponse`, à côté de `routingInfo`, en JS `doc.ready = st.ready; routed(st);` ; adapter `internal/registry_test.go:627`, `:835`.
- Verdict : corrigé.
- Fix : `readyResponse` porte `ready` (un `readyState`) à côté de `routingInfo` ; la page fait `doc.ready = st.ready`. `class`. Commits `3740c7d`.

<a id="w-finding-014"></a>
### W-FINDING-014 · Le log de l'instance B n'est jamais affiché par `run.sh`

- [code-quality] Medium cx:S, origine `new` (`715401f`), `scripts/e2e/instance-routing.js:80`.
- Cas : un check du run 013 échoue après le démarrage de B → `dump_logs` (`scripts/e2e/run.sh:106`) affiche `server.log` (A) et `chrome.log`, sans une ligne de B. La sortie de B reste dans la variable `out`, lue seulement si B ne démarre pas.
- Preuve : lecture de `b.stdout.on("data", (d) => { out += d; ... })` ; `dump_logs` lit `$tmp/server.log` et `$tmp/chrome.log` seulement.
- Correctif proposé (inst) : lancer B avec sa sortie dans `<work>/b.log`, lire la ligne de démarrage dans ce fichier, ajouter `--- instance B log` à `dump_logs`.
- Verdict : corrigé.
- Fix : B écrit dans `<work>/b.log` ; `dump_logs` l'affiche. `inst`. Commits `31ebe67`.

<a id="w-finding-015"></a>
### W-FINDING-015 · L'attente « Sent → :port » écrite sept fois dans `manual-trigger.js`

- [code-quality] Medium cx:S, origine `new` (`715401f`), `scripts/e2e/manual-trigger.js:162`.
- Cas : l'attente `textContent === "Sent → :${port}"` est écrite 7 fois, plus la variante de la live region (`:204`). Un nouveau libellé change 8 lignes ; une ligne oubliée échoue en timeout de 4 s, sans message clair.
- Preuve : `grep -c 'Sent → :\${port}' manual-trigger.js` → 7.
- Gravité corrigée : Low → Medium (doctrine, Medium : une valeur écrite à plusieurs endroits).
- Correctif proposé (inst) : `const SENT = "Sent → :" + port;` et un `waitSent()` utilisé aux 8 endroits.
- Verdict : corrigé.
- Fix : `SENT` et `waitSent()` dans `manual-trigger.js`. `inst`. Commits `b32ff6d`.

<a id="w-finding-016"></a>
### W-FINDING-016 · Pas de test du flock exclusif de `--claim`

- [tests] Medium cx:M, origine `new` (`2b67210`), `internal/pending.go:130`.
- Cas : le spec dit « under the flock of the sidecar, checks again … then writes `pid` ». Mutant `exclusive = false` : deux `--claim` concurrents, ou un `--claim` pendant un Go (`internal/ready.go:153`), écrivent tous deux `pid` ; la suite reste verte.
- Preuve : worktree jetable sur HEAD, mutant `exclusive = false`, `go test -count=1 ./internal/ -run 'TestPending|TestFindSidecars'` → `ok`; `go test -count=1 ./cmd/` → `ok`.
- Correctif proposé (class) : test dans `internal/pending_test.go` : tenir `lockSidecar(pf.sidecar, os.O_RDONLY, false)`, lancer `--claim` dans une goroutine, vérifier qu'il ne rend pas avant `releaseSidecar`, puis qu'il écrit `pid`.
- Verdict : fix.
- Fix : `TestPendingClaimWaitsForTheLock` : un verrou partagé tenu, `--claim` ne rend pas avant `releaseSidecar`, puis écrit `pid` ; échoue avec le mutant `exclusive = false`. `class`. Commits `dec55a5`.

<a id="w-finding-017"></a>
### W-FINDING-017 · Pas de test du re-contrôle du trigger manuel par `--claim`

- [tests] Medium cx:S, origine `new` (`2b67210`), `internal/pending.go:166`.
- Cas : le spec dit « `--claim` … checks again that the trigger is manual … exit 1 and nothing written otherwise ». Mutant sans `if claim != 0 { return nil, nil }` : `--claim` sur un sidecar en auto imprime les entrées et sort 0 ; aucun test ne passe au rouge.
- Preuve : même worktree et même commande que [W-FINDING-016](#w-finding-016), mutant combiné → `ok` sur `./internal/` et `./cmd/`.
- Correctif proposé (inst) : dans `TestPendingClaim`, un cas `pf.write(triggerAuto)` avec un ready file de A non pris : `found == false`, sortie vide, ready file inchangé.
- Verdict : fix.
- Fix : cas `triggerAuto` dans `TestPendingClaim` : refus, sortie vide, ready file inchangé ; échoue sans le re-contrôle. `inst`. Commits `dec55a5`.

<a id="w-finding-018"></a>
### W-FINDING-018 · Pas de test du 400 sur un corps de POST ready illisible

- [tests] Medium cx:S, origine `new` (`2b67210`), `internal/ready.go:254`.
- Cas : POST `/__annotations/ready` avec le corps `{"via":` → 400 attendu, aucun fichier ready. Sans la vérification, Go part vers la première cible et écrit le fichier ready ; aucun test ne le voit.
- Preuve : worktree jetable sur `b5a576a`, `decodeJSONBody(...)`/400 remplacé par `_ = json.NewDecoder(r.Body).Decode(&body)`, `go test -count=1 ./internal -run 'Ready|Go|Via|Restamp|Shared|Annotation|Merge|Pending|Sidecar'` → `ok`.
- Correctif proposé (class) : un cas de `TestReadyRefusals` : corps JSON tronqué → 400 et `readyContent() == ""`.
- Verdict : nothing : test d'un cas limite (la page envoie toujours `JSON.stringify`), sous High.

<a id="w-finding-019"></a>
### W-FINDING-019 · Pas de test du port choisi envoyé par Go

- [tests] Medium cx:S, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:377`.
- Cas : le spec dit « Go (`POST /__annotations/ready`) takes the chosen port ». Sans `init.body = JSON.stringify({ via: chosen() })`, le serveur prend la première cible = soi ; `manual-trigger.js` lance une seule instance, `instance-routing.js` ne clique jamais Go : l'e2e reste vert.
- Preuve : `grep -n "annotation-go" scripts/e2e/instance-routing.js` → aucune ligne.
- Correctif proposé (inst) : dans `instance-routing.js`, B choisi et page en manual, cliquer Go ; vérifier `via` = B dans le ready file et « Sent → :B ».
- Verdict : corrigé.
- Fix : étape e2e : Go avec B choisi, ready file à `via` = B et « Sent → :B ». `inst`. Commits `7d77e57`.

<a id="w-finding-020"></a>
### W-FINDING-020 · Pas de test de la bannière redessinée par `poll`

- [tests] Medium cx:S, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:337`.
- Cas : le spec dit « The page redraws the banner on each fetch, `poll` included ». La branche `routed(d); drawBanner()` de `poll()` ne tourne qu'en manual avec l'état `working` ; `instance-routing.js` tourne en auto. Sans ces deux appels, l'e2e reste vert.
- Preuve : `grep -n "manual\|trigger" scripts/e2e/instance-routing.js` → aucune ligne.
- Correctif proposé (inst) : dans `instance-routing.js`, page en manual avec l'état `working`, poussée socket bloquée, `targets` changé ; vérifier la bannière en moins d'une période de `poll` (5 s).
- Verdict : corrigé.
- Fix : étape e2e : page en manual, lot `working`, événement du socket bloqué, B ouvre la page : bannière par `poll`. `inst`. Commits `ba7d818`.

<a id="w-finding-021"></a>
### W-FINDING-021 · Pas de test de l'absence de flèche sur un message `via` = `self`

- [tests] Medium cx:S, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:512`.
- Cas : le spec dit « Each reader message whose `via` is not `self` shows "→ :<via>" ». Sans le test `m.via === doc.self` dans `routeOf`, chaque message d'une page à une instance affiche « → :<port> » ; l'e2e reste vert.
- Preuve : `grep -n "annot-msg-reader\|small" scripts/e2e/manual-trigger.js` → aucune ligne.
- Correctif proposé (inst) : dans `instance-routing.js`, un message avec A choisi, dont le libellé ne contient pas « → ».
- Verdict : corrigé.
- Fix : étape e2e : messages routés vers A sans « → ». `inst`. Commits `8ac24e9`.

<a id="w-finding-022"></a>
### W-FINDING-022 · Pas de test du repli d'un port mémorisé hors de `targets`

- [tests] Medium cx:S, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:212`.
- Cas : le spec dit « a stored port no longer in `targets` falls back to the first of `targets` ». Si `chosen()` rend le port mémorisé sans contrôle, B redémarré réapparaît choisi ; aucun check ne tourne après « message shows (gone) ».
- Preuve : `instance-routing.js` s'arrête au check « message shows (gone) » ; aucun message ni Go après l'arrêt de B.
- Correctif proposé (inst) : après l'arrêt de B, `sessionStorage` à B, un message ; vérifier que le corps du PUT nomme A.
- Verdict : corrigé.
- Fix : étape e2e : B mémorisé après son arrêt, le PUT nomme A. `inst`. Commits `c4007ff`.

<a id="w-finding-023"></a>
### W-FINDING-023 · « choice survives a reload » passe sans rechargement

- [tests] Medium cx:S, origine `new` (`715401f`), `scripts/e2e/instance-routing.js:149`.
- Cas : le check attend `sleep(500)` puis une condition que l'ancienne page remplit déjà (bannière affichée). Un `Page.reload` lent de plus de 500 ms → PASS sur l'ancien document, sans test de la restauration depuis `sessionStorage`. Le check n'affirme pas `marker === null`.
- Preuve : `scripts/e2e/instance-routing.js:148-151` ; `controls-b5a576a.log:96` montre `marker:null`, que le test ne vérifie pas.
- Correctif proposé (inst) : remplacer le sleep par `waitFor("!window.__marker && document.readyState === 'complete' && …")` et affirmer `marker === null`.
- Verdict : corrigé.
- Fix : attente de `!window.__marker` au lieu du sleep, `marker === null` affirmé. `inst`. Commits `5b521c7`.

<a id="w-finding-024"></a>
### W-FINDING-024 · `contains(root, key)` : nom générique à côté de `slices.Contains`

- [code-quality] Low cx:S, origine `new` (`9845bb3`), `internal/registry.go:94`.
- Cas : `contains(in.Root, key) && slices.Contains(in.Pages, key)` (`:198`) : deux « contains » de sens différent sur une ligne.
- Preuve : `internal/registry.go:198`.
- Correctif proposé (inst) : renommer en `under(root, key)`.
- Verdict : corrigé.
- Fix : `contains` renommé `under`. `inst`. Commits `b8a5371`.

<a id="w-finding-025"></a>
### W-FINDING-025 · Lignes de plus de 100 colonnes dans `README.md` et `internal/pending.go`

- [code-quality] [docs] Low cx:S, origine `fix` (`098fa7d`), `internal/pending.go:87`.
- Cas : après le rewrap de `ea1c44f` (limite de 100 posée par R1-26 et R2-04), restent au-delà : `internal/pending.go:87` (106 col., commentaire de `findSidecars`, ajouté par `098fa7d`), `:92` (101) ; `README.md:173` (105), `:175` (101), `:179` (102), `:241` (102).
- Preuve : `awk '{gsub(/\t/,"    "); if (length($0)>100 && $0 ~ /^ *\/\//) print FILENAME":"FNR}' internal/*.go` → `internal/pending.go:87`, `:92` ; `git diff -U0 e24fad1...b5a576a -- README.md` filtré sur len>100 → 4 lignes de prose.
- Correctif proposé (inst) : rewrap de ces lignes.
- Verdict : corrigé.
- Fix : rewrap de `internal/pending.go` et des lignes de prose de `README.md` de la branche. `inst`. Commits `ff00119`.

<a id="w-finding-026"></a>
### W-FINDING-026 · Champ de table `live2` dans `TestPeers`

- [code-quality] Low cx:S, origine `new` (`9845bb3`), `internal/registry_test.go:94`.
- Cas : `live2` contient le résultat attendu de `ports()` ; `live` est déjà le nom de l'entrée. Un lecteur y voit une seconde liste d'instances.
- Preuve : `internal/registry_test.go:123` `p.ports(); !reflect.DeepEqual(got, tc.live2)`.
- Correctif proposé (inst) : renommer en `ports`.
- Verdict : corrigé.
- Fix : `live2` renommé `ports`. `inst`. Commits `933d305`.

<a id="w-finding-027"></a>
### W-FINDING-027 · `regLive` et `liveFor` : points d'injection jamais variés

- [code-quality] Low cx:S, origine `new` (`9845bb3`), `internal/server.go:71`.
- Cas : `Server.regLive` et `registry.liveFor` valent toujours `registryLive` ; `Pending.owns` lit la constante. Un lecteur croit la durée réglable par instance ; un réglage du serveur ne changerait pas `pending`.
- Preuve : `grep -rn "regLive\|liveFor" internal cmd` → `server.go:71`, `:122`, `:283` et `registry.go` seulement.
- Correctif proposé (inst) : supprimer `regLive` et `liveFor`, `registryLive` dans `registry.read`.
- Verdict : corrigé.
- Fix : `regLive` et `liveFor` supprimés, `registryLive` dans `readWith`. `inst`. Commits `e152fbc`.

<a id="w-finding-028"></a>
### W-FINDING-028 · `restamp` refait l'accès au dernier message

- [code-quality] Low cx:S, origine `new` (`2b67210`), `internal/annotations.go:409`.
- Cas : `restamp` refait `&a.Thread[len(a.Thread)-1]` à côté de `lastVia` (`internal/ready.go:131`) et `isOpen` (`:145`), dans deux fichiers.
- Preuve : lecture de `internal/annotations.go:404-414`, `internal/ready.go:130-145`.
- Correctif proposé (inst) : un `lastMessage(a *annotation) *message` dans `annotations.go`, utilisé par `isOpen`, `lastVia` et `restamp`.
- Verdict : corrigé.
- Fix : `lastMessage` utilisé par `isOpen`, `lastVia`, `restamp`. `inst`. Commits `a8029e1`.

<a id="w-finding-029"></a>
### W-FINDING-029 · `NewServer` : deux chaînes adjacentes `settingsPath, registryDir`

- [code-quality] Low cx:S, origine `new` (`9845bb3`), `internal/server.go:117`.
- Cas : `NewServer(..., registryDir, settingsPath)` compile et écrit les réglages dans le répertoire du registre.
- Preuve : lecture de `internal/server.go:117` et des appels de test (`NewParser(), "", ""`).
- Correctif proposé (inst) : `registryDir` par un champ posé après construction, comme `signals` et `regEvery`.
- Verdict : corrigé.
- Fix : `registryDir` posé par `Server.UseRegistry`, retiré de `NewServer`. `inst`. Commits `87be25b`.

<a id="w-finding-030"></a>
### W-FINDING-030 · `routeOf` lit `doc.live` sans accesseur, à côté de `targets()`

- [code-quality] Low cx:S, origine `new` (`f6a9479`), `defaults/static/js/annotate.js:513`.
- Cas : deux façons de lire un champ de routage qui peut manquer : `targets()` (`:203`) et `(doc.live || [])` en ligne.
- Preuve : `grep -n "doc.live\|targets()" annotate.js` → 513, 203-213.
- Correctif proposé (inst) : `function live() { return doc.live || []; }`, utilisé par `routeOf`.
- Verdict : corrigé.
- Fix : accesseur `live()` utilisé par `routeOf`. `inst`. Commits `3a8fd76`.

<a id="w-finding-031"></a>
### W-FINDING-031 · En-tête du sidecar : « in manual only once claimed » ambigu

- [docs] Low cx:S, origine `fix` (`cbf8bbb`), `internal/annotations.go:58`.
- Cas : l'en-tête dit « handle only the entries it lists … in manual only once claimed (below) ». En manual, `pending` liste les entrées avec `ready` tant que le lot n'est pas pris, et rien après la prise (`pendingLines` : `rf.PID != nil` rend nil). Un agent qui relance `pending` après sa prise reçoit exit 1. Le paragraphe suivant le dit (« Handle the entries the claim lists ») : ambigu, pas faux.
- Preuve : `internal/pending.go:176` `if err != nil || rf.PID != nil || !owns(key, rf.Via) { return nil, nil }` (lu, non exécuté).
- Correctif proposé (inst) : « in manual, the entries of a `ready` batch, once you claimed it (the claim lists them) ».
- Verdict : corrigé.
- Fix : en-tête : « In manual, handle the entries of a ready batch once you claimed it: the claim lists them ». `inst`. Commits `b42b448`.

<a id="w-finding-032"></a>
### W-FINDING-032 · Variable `asleep` pour le cas « pas de fichier, PID vivant »

- [tests] Low cx:S, origine `new` (`2b67210`), `internal/pending_test.go:95`.
- Cas : `asleep` nomme le PID du cas « routing off: no file, live PID » ; « asleep » est le nom que le spec donne au cas inverse de AC12 (fichier présent, `seen_at` vieux, rien n'est possédé).
- Preuve : `internal/pending_test.go:95` `asleep := livePID(t)` ; `:107`.
- Correctif proposé (inst) : renommer en `noFile`.
- Verdict : fix.
- Fix : `asleep` renommé `noFile`. `inst`. Commits `1952c83`.

<a id="w-finding-033"></a>
### W-FINDING-033 · Fichiers temporaires `.<pid>.json-*` jamais purgés

- [resiliency] [outside the level] Low cx:S, origine `new` (`9845bb3`), `internal/registry.go:245`.
- Cas : un kill entre `CreateTemp` et `Rename` dans `replaceFile` laisse `.<pid>.json-*` dans le répertoire du registre ; `scanRegistry` saute tout nom autre que `<pid>.json`. Sans effet sur les lectures.
- Preuve : [unverified], lecture de `scanRegistry` et `replaceFile`.
- Correctif proposé (inst) : `scanRegistry` retire un `.<pid>.json-*` dont le PID est mort.
- Verdict : nothing : hors niveau (outil local personnel : resiliency limitée aux menaces de la liste « always the user's »).

## Tours

| Tour | Périmètre | Angles lancés | Critical | High | Medium | Low | Hors niveau | fix | new | old | Angles actifs ensuite |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `e24fad1..b5a576a` | functional, tests, code-quality × M1 M2 M3 ; security, resiliency, docs | 0 | 2 | 21 | 9 | 1 | 4 | 29 | 0 | functional |
