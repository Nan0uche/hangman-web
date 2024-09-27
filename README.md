# Jeu du Pendu en Go

Bienvenue dans le jeu du Pendu implémenté en Go ! Ce jeu web simple permet aux joueurs de deviner un mot et d'éviter d'être "pendus".

## Mise en route

### Prérequis

Assurez-vous d'avoir Go installé sur votre machine. Si ce n'est pas le cas, vous pouvez le télécharger et l'installer depuis [ici](https://golang.org/dl/).

### Installation

1. Clonez le dépôt :

```bash
   git clone https://ytrack.learn.ynov.com/git/dnathan/hangman-web.git
```
2. Assurez-vous d'être dans le répertoire du projet :
```bash
    cd (Emplacement du projet)
```
3. Exécutez l'application Go :
```bash
    go run main.go
```
## Comment jouer
- Visitez http://localhost:8080 dans votre navigateur web.

- Vous serez invité à entrer votre nom. Une fois saisi, cliquez sur "Suivant".

- Choisissez le niveau de difficulté (Facile, Normal, Difficile) pour commencer une nouvelle partie.

- Devinez les lettres ou le mot entier pour révéler le mot caché.

- Continuez à jouer jusqu'à ce que vous deviniez correctement le mot ou que vous épuisiez vos tentatives.
## Fonctionnalités
- Sauvegarde et Continuation : Le jeu sauvegarde votre progression et de continuer plus tard.

- Niveaux de difficulté : Choisissez parmi trois niveaux de difficulté, chacun avec un ensemble différent de mots.

- Design Responsive : L'interface web est conçue pour fonctionner bien sur différents appareils.
## Structure des fichiers
- **main.go** : Le fichier principal de l'application qui configure le serveur HTTP et gère le routage.

- **templates/** : Contient les modèles HTML pour différentes pages (accueil, saisie du nom, sélecteur de difficulté, jeu, fin de partie).

- **picture/** : Inclut des images représentant le dessin du pendu à différentes étapes.

## Crédits
CHIPPEY Théo : https://ytrack.learn.ynov.com/git/chtheo  
ROUSSEL Mathéo : https://ytrack.learn.ynov.com/git/rmatheo  
DERC Nathan : https://ytrack.learn.ynov.com/git/dnathan  
Projet réalisé dans le cadre d'une soutenance par groupe à faire pour le Vendredi 22 Décembre 2023 chez Ynov Bordeaux Campus, classe B1 Informatique.

Plus d'information sur la stucture Ynov Bordeaux Campus : https://ynov-bordeaux.com/