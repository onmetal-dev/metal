import hetznerLocations from "./hcloud/locations";

export function networkZoneForLocation(location: string) {
  const l = hetznerLocations.locations.find((l) => l.name === location);
  if (!l) {
    throw new Error("Invalid location.");
  }
  return l.network_zone;
}
