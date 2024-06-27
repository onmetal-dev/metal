import { components } from "./index";
type schemas = components["schemas"];
import hetznerLocationsData from "./locations.json";
const hetznerLocations: schemas["list_locations_response"] =
  hetznerLocationsData as schemas["list_locations_response"];

export default hetznerLocations;
