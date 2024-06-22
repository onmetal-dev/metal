import { components } from "./index";
type schemas = components["schemas"];
import hetznerLocationsData from "./locations.json";
const hetznerLocations: schemas["list_locations_response"] =
  hetznerLocationsData as schemas["list_locations_response"];

export default hetznerLocations;

export function networkZoneForLocation(location: string) {
  const l = hetznerLocations.locations.find((l) => l.name === location);
  if (!l) {
    throw new Error("Invalid location.");
  }
  return l.network_zone;
}
