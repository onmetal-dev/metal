import { components } from "./index";
type schemas = components["schemas"];
import hetznerDatacentersData from "./datacenters.json";
const hetznerDatacenters: schemas["list_datacenters_response"] =
  hetznerDatacentersData as schemas["list_datacenters_response"];

export default hetznerDatacenters;
