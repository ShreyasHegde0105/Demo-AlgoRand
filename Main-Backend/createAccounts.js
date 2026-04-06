const algosdk = require("algosdk");

function createAccount(role) {
  const acc = algosdk.generateAccount();
  console.log(`\n=== ${role} ===`);
  console.log("Address:", acc.addr.toString()); // ✅ FIX HERE
  console.log("Mnemonic:", algosdk.secretKeyToMnemonic(acc.sk));
}

createAccount("CLIENT");
createAccount("FREELANCER");
createAccount("REVIEWER");