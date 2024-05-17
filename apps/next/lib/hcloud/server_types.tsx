import { components } from "./index";
type schemas = components["schemas"];
import hetznerServerTypesData from "./server_types.json";
const hetznerServerTypes: schemas["list_server_types_response"] =
  hetznerServerTypesData as schemas["list_server_types_response"];

export default hetznerServerTypes;
