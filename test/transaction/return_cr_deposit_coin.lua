-- Copyright (c) 2017-2019 Elastos Foundation
-- Use of this source code is governed by an MIT
-- license that can be found in the LICENSE file.
--

local m = require("api")

-- client: path, password, if create
local wallet = client.new("keystore.dat", "123", false)

-- account
local addr = wallet:get_address()
local pubkey = wallet:get_publickey()

print(addr)
print(pubkey)

-- assetID
local assetID = m.get_asset_id()

-- amount, fee
local amount = 0.199
local fee = 0.001
local recipient = "EJMzC16Eorq9CuFCGtyMrq4Jmgw9jYCHQR"
local deposit_addr = "DgauQGNDXYkn3xvVhHHDy4utmoQUjCdJgM"

-- return deposit payload
local rp_payload = returndepositcoin.new()
print(rp_payload:get())

-- transaction: version, txType, payloadVersion, payload, locktime
local tx = transaction.new(9, 0x24, 0, rp_payload, 0)

-- input: from, amount + fee
local charge = tx:appendenough(deposit_addr, (amount + fee) * 100000000)
print(charge)

-- default output payload
local default_output = defaultoutput.new()

-- output: asset_id, value, recipient, output_paload_type, output_paload
local amount_output = output.new(assetID, amount * 100000000, recipient, 0, default_output)
tx:appendtxout(amount_output)
if (charge ~= 0)
then
    local charge_output = output.new(assetID, charge, deposit_addr, 0, default_output)
    tx:appendtxout(charge_output)
end
-- print(amount_output:get())

-- sign
tx:sign(wallet)
print(tx:get())

-- send
local hash = tx:hash()
local res = m.send_tx(tx)

print("sending " .. hash)

if (res ~= hash)
then
    print(res)
else
    print("tx send success")
end