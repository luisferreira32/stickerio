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
- [x] Experiment with generics for the repetivive code with different types;
- [x] Implement event insertion API endpoints (movement, upgrades, training, etc.);
- [x] Implement event sourcing/handling;
- [ ] Implement CLI to obtain the views;
- [ ] Implement CLI to submit events;
- [ ] Introduce some API e2e testing;
- [ ] Allow setting the type of movement the troops should do (attack vs. reinforce/relocate);
- [ ] Balance game configurations for a decent playing experience;

### Documentation
- [x] Add a LICENSE
- [ ] Extend README.md with relevant info about contribution, roadmap, purposes, etc.;
- [ ] Define guidelines for CONTRIBUTE.md;
- [ ] Formalize architecture decisions in ARCHITECTURE.md;

### User management
- [ ] Create a user database;
- [ ] Generate and sign tokens;
- [ ] Validate tokens on every API requests and pre-populate context with player info;

## Alpha version 1.0

### Gameplay
- [ ] Add units with different power dynamics;
- [ ] Extend resource types;
- [ ] Add "premium" resource that is unobtainable;
- [ ] Extend current CLI with fancier text outputs;
- [ ] Ensure current CLI can load different languages;

### User management
- [ ] Allow federated sign-in (e.g., google, facebook, etc.);

### Game implementation
- [ ] Re-work efficiency of the event sourcing (serialization, etc.);
- [ ] Introduce API rating;

## Alpha version 2.0

### Gameplay
- [ ] In-game mailing system;
- [ ] Implement a terminal based UI;
- [ ] Allow minig of "premium" resource and allow plugin extension for that;
