# Go Distributed File System

GDFS est un système de stockage de fichier décentralisé qui repose sur une implémentation de table de hachage distribuée en Go. Chaque noeud peut stocker et retrouver des fichiers sur son réseau pair-à-pair, et profite d’une grande résilience faces aux pannes grâce aux mécanismes de réplication des données et de mise à jour continue des tables de routage.

## Utilisation

### Démarrer un noeud

```bash
go cmd/node/main.go # Démarre un noeud initial écoutant sur le port 42042
```

| Option       | Requise | Description                                          |
|--------------|---------|------------------------------------------------------|
| `-port`      | non     | Port d'écoute du noeud (par défaut 42042)            |
| `-bootstrap` | non     | Adresse d'un noeud existant pour rejoindre un réseau |

### Stocker et retrouver un fichier

```bash
# Stocke un fichier sur un réseau
go cmd/cli/main.go -store -file {chemin} -addr {adresse}

# Retrouve un fichier sur le réseau et le stocke à {chemin}
go cmd/cli/main.go -find -id {identifiant} -file {chemin} -addr {adresse} 
```

`addr` est l'adresse d'un noeud du réseau (par défaut: 127.0.0.1:42042).

## Configuration par défaut

| Constante        | Valeur par défaut | Description                                      |
|------------------|-------------------|--------------------------------------------------|
| IdSize           | 20                | Taille des identifiants en octet                 |
| ValueSize        | 1024              | Taille d'une valeur en octet                     |
| maxReplicasCount | 5                 | Nombre maximum de replicas pour une valeur       |
| storageTtl       | 60 minutes        | Durée de vie d'une valeur dans le stockage local |
| storageCapacity  | 65 536            | Nombre maximum de valeurs stockées localement    |

Toutes les constantes de configuration se trouvent dans `core/config.go`.

## Tester

```bash
go test -v ./testing
```
