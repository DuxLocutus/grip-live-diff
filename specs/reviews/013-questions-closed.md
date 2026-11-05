# Questions fermées · spec 013

<a id="013-question-01"></a>
## 013-QUESTION-01 · Instance vivante dont le battement échoue · close le 2026-11-05

- Décide : l'utilisateur.
- Source : revue informelle de la branche, finding R1-11 (`internal/registry.go`).
- Contexte : le spec dit « its file is there but not live (stopped, asleep) … nothing is owned ». Une instance vivante qui n'arrive plus à réécrire son fichier a un fichier périmé : `pending` dit qu'elle ne possède rien, alors que son serveur écrit encore `via` = son port. Ses messages vont au repli (port vivant le plus bas qui contient le fichier), ou à personne.
- Cas : registre en lecture seule après le démarrage, disque plein.
- Proposition : (b) le serveur cesse d'écrire `via` quand son battement échoue au-delà de 60 s (routage désactivé), comme quand le registre est illisible au démarrage. Autre option : (a) `pending` traite un fichier périmé avec PID vivant comme « pas de fichier » et l'instance possède tous les fils ouverts.
- Échanges : l’utilisateur choisit (b), 2026-11-05. Appliqué : le serveur coupe le routage quand son battement n’est plus sur disque depuis 60 s (`internal/registry.go`, `readWith`), test `TestStaleHeartbeatTurnsRoutingOff`.
- Clôture : option (b).
