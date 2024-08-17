/* resend of a fulfillment message */
SELECT * from pgmq.send(
  queue_name  => 'fulfillment',
  msg         => '{
  "TeamId": "team_01j422npwzez4amhhg45tnsqp1",
  "UserId": "user_01j422mtm7ez4awehzg3wdwbkz",
  "CellName": "",
  "DnsZoneId": "54807646b317d33ac57a02ac7887ff72",
  "LocationId": "HEL1",
  "OfferingId": "AX41-NVMe",
  "StepServerId": "server_01j54jsrnvek6bmwmednc5ztvp",
  "StepTalosOnline": false,
  "StepServerOnline": true,
  "StepAddServerToCell": false,
  "StepPaymentReceived": true,
  "StepServerInstalled": false,
  "StepProviderServerId": "2428820",
  "StripeCheckoutSessionId": "cs_test_a1zvgbwvVcuoWJUTwQDjCMEgei36aNGXZyItPMEfXrMB2esLM5Jwewih8U",
  "StepProviderTransactionId": "B20240813-2900159-2517363"
}');