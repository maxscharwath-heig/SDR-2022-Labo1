# SDR 2022 / Labo 1 - Programmation répartie

> Nicolas Crauaz & Maxime Scharwath

Laboratoire 1

## Installation

Pour démarrer l'
`go get -d`

## Configuration

La configuration du serveur et du client sont séparées dans deux fichiers différents.

La configuration du serveur se trouve dans `config/server.json`: 

```json
  "host": "localhost",    // IP / nom DNS du serveur
  "port": 9000,           // Port d'écoute du serveur
  "users": [...],         // Utilisateurs enregistrés
  "events": [...]         // Evénements enregistrés
```

La configuration du client se trouve dans `config/client.json`:

```json
{
  "srvHost": "localhost",  // IP / nom DNS du serveur
  "srvPort": 9000          // Port d'écoute du serveur
}
```

## Utilisation

### Lancer le serveur (directement, ou via un exécutable)

> `go run server.go`
ou
> `go build server.go && ./server`

Le serveur attendra des connexions sur le port TCP configuré dans `config/server.json`

### Lancer un client (directement, ou via un exécutable)

> `go run client.go`
ou
> `go build client.go && ./client`

### Liste de commandes disponible

#### Créer une manifestation

> create

Puis saisir les informations demandées. Il est nécessaire de s'authentifier pour créer une manifestation.

#### Clôturer une manifestation

> close

Puis saisir les informations demandées.
Il est nécessaire de s'authentifier et d'être le créateur de la manifestation pour la clôturer.

#### Inscription d’un bénévole

> register

Puis saisir les informations demandées. Il est nécessaire de s'authentifier.
Il est possible d'être inscrit qu'à un seul poste par manifestation, l'inscription la plus récente sera conservée.

#### Liste des manifestations

> show

Affiche l'état de toutes les manifestations.

// TODO: Mettre une capture

#### Informations d'une manifestation

> show `<numéro manifesation>`

Affiche les informations d'une manifestation

// TODO: Mettre une capture

#### Répartition des postes pour une manifestation

> show `<numéro manifesation>` --resume

Affiche l'état des postes d'une manifestation.

// TODO: Mettre une capture

## Limitations

Il n'y a pas de persistance des données au-delà de l'exécution du serveur.

## Tâches

- [x] Fichier de configuration (port, utilisateurs (organisateur ou bénévole), manifestation)
- [x] CLI (client)
- [x] Message de bienvenue et d'aide (client)
- [x] Séléction de l'étape
    - [x] Créer une manifestation (nom, username, password, nom des postes et nombre de bénévoles)
    - [x] Cloturer une manifestation (numéro, username, password)
- [ ] Inscription à une manifestation (username, password, numéro manif et numéro du poste)
    - [ ] Vérifier que le poste existe
    - [ ] Vérifier que le poste n'est pas déjà pris
    - [ ] Vérifier que le nombre max de bénévoles n'est pas atteint
    - [ ] Si l'utilisateur s'incrit plusieurs fois, seule la dernière est gardée

Pas d'auth sur le listing

- [x] Lister toutes les manifestations (affiche: numéro, nom, nom organisateur, ouvert (oui /non))

- [x] Lister une manifestation (par numéro): afficher tous les postes (numéros, nom et nombre max de bénévoles)

- [ ] Afficher tableau avec en ligne les noms des bénévoles ayant répondu et en colonne les
  numéros des postes prévus. Les cases correspondant aux inscriptions de bénévoles sur des postes seront marquées. Les
  autres seront laissées à blanc. En haut de colonne, le nombre de bénévoles attendus sera affiché. 