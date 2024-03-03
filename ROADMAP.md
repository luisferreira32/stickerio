# Roadmap

Here lies the _detailed_ roadmap of Stickerio. If you are interested in the project feel free to contribute following the guidelines of [CONTRIBUTE.md](./CONTRIBUTE.md).

## Pre-release

### Game implementation
- [x] Initial project skeletton;
- [x] Initial table schema;
- [x] Simple view endpoint of a single city;
- [ ] Extend views for a full basic possibility (city view, movements view, upgrade buildings/training units view, map view);
- [ ] Allow training of units;
- [ ] Allow movement of units with basic output: if arriving at a player owned city they change garrisions, if arriving at a non-player owned city they attack, if arriving at an empty map location they forage a random (low) quantity of resources;
- [ ] Allow upgrade of buildings: military buildings decrease training time, resource buildings increase resources per second;
- [ ] Implement text-based CLI controls for getting the info
- [ ] Implement text-based CLI controls for actions: training, movement, upgrading.
- [ ] Balance configurations

### Documentation
- [x] Add a LICENSE
- [ ] Extend README.md with relevant info about contribution, roadmap, purposes, etc.
- [ ] Define guidelines for CONTRIBUTE.md
- [ ] Formalize architecture decisions in ARCHITECTURE.md

### User management
- [ ] Create a user database
- [ ] Generate and sign tokens
- [ ] Validate tokens on every API requests and pre-populate context with player info

## Alpha version 1.0

### Gameplay
- [ ] Add units with different power dynamics
- [ ] Extend resource types
- [ ] Extend current CLI with fancier text outputs

### User management
- [ ] Allow federated sign-in (e.g., google, facebook, etc.)

## Alpha version 2.0

### Gameplay
- [ ] In-game mailing system
- [ ] Implement a terminal based UI