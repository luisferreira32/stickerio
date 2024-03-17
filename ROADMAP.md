# Roadmap

Here lies the _detailed_ roadmap of Stickerio. If you are interested in the project feel free to contribute following the guidelines of [CONTRIBUTE.md](./CONTRIBUTE.md).

## Pre-release

### Game implementation
- [x] Initial project skeletton;
- [x] Initial table schema;
- [x] Simple view endpoint of a single city;
- [x] Implement view API endpoints;
- [x] Migrate API configuration to a more standardized tool (e.g., proto / OAS);
- [x] Better separate business logic from API conversion;
- [ ] Experiment with generics for the repetivive code with different types;
- [ ] Implement event insertion API endpoints (movement, upgrades, training, etc.);
- [ ] Implement event sourcing/handling;
- [ ] Implement CLI to obtain the views;
- [ ] Implement CLI to submit events;
- [ ] Balance game configurations for a decent playing experience;

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

### Game implementation
- [ ] Re-work efficiency of the event sourcing (serialization, etc.)

## Alpha version 2.0

### Gameplay
- [ ] In-game mailing system
- [ ] Implement a terminal based UI