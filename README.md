# SDR-Labo1
Laboratoire 1

## Installation

`go get -d`


## Utilisation

Lancer le serveur:

> `go run server.go`

## Tâches

- [x] Fichier de configuration (port, utilisateurs (organisateur ou bénévole), manifestation)
- [x] CLI (client)
- [x] Message de bienvenue et d'aide (client)
- [ ] Séléction de l'étape
   - [ ] Créer une manifestation (nom, username, password, nom des postes et nombre de bénévoles)
   - [ ] Cloturer une manifestation (numéro, username, password)
- [ ] Inscription à une manifestation (username, password, numéro manif et numéro du poste)
   - [ ] Vérifier que le poste existe
   - [ ] Vérifier que le poste n'est pas déjà pris
   - [ ] Vérifier que le nombre max de bénévoles n'est pas atteint
   - [ ] Si l'utilisateur s'incrit plusieurs fois, seule la dernière est gardée

Pas d'auth sur le listing

- [x] Lister toutes les manifestations (affiche: numéro, nom, nom organisateur, ouvert (oui /non))

- [ ] Lister une manifestation (par numéro): afficher tous les postes (numéros, nom et nombre max de bénévoles)

- [ ] Afficher tableau avec en ligne les noms des bénévoles ayant répondu et en colonne les
numéros des postes prévus. Les cases correspondant aux inscriptions de bénévoles sur des postes seront marquées. Les autres seront laissées à blanc. En haut de colonne, le nombre de bénévoles attendus sera affiché. 