import { components } from "./index";
type schemas = components["schemas"];
import hetznerPricingData from "./pricing.json";
const hetznerPricing: schemas["list_prices_response"] =
  hetznerPricingData as schemas["list_prices_response"];

export default hetznerPricing;
