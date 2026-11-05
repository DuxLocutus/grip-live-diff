# Cas · spec 013

| Id | Cas | Ce que l'agent a fait | Règle appliquée | Lien |
|---|---|---|---|---|
| <a id="w-case-001"></a>W-CASE-001 | Lot `working` dont le `via` est parti : le spec dit que Go réécrit `via` pour un lot « sent or worked on », la page ne fixe pas quand Go est actif. | Go actif seulement sur un lot `sent` dont le `via` n'est pas dans `live` ; un lot `working` garde Go désactivé (son agent vit encore ; ✕ reste offert). | Table des niveaux, « Visible behaviour no document states » : option la plus simple qui ne perd pas de données. | [W-FINDING-001](013-instance-routing-review-2026-11-05-detailed.md#w-finding-001) |
