export type NixpackPlan = {
  phases: {
    build: NixpackPhase
    setup: NixpackSetup
    install: NixpackPhase
  }
}

export type NixpackPhase = {
  cacheDirectories: string[]
  cmds: string[]
  dependsOn: string[]
}

export type NixpackSetup = {
  nixPkgs: string[]
  nixLibs: string[]
  nixOverlays: string[]
  nixpkgsArchive: string
  aptPkgs: string[]  
}
