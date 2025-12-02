import * as KeetaNet from '@keetanetwork/keetanet-client';

const usernameTokenAddress = 'keeta_ambae3744pa4jpztc3fourfaanw3prbwoltne3jinondcx6kw62vtsrceko6i';
const CID = 'QmcniBv7UQ4gGPQQW2BwbD4ZZHzN3o3tPuNLZCbBchd1zh';

const burnPublicKey = 'keeta_aeaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaazpi2nodu';

async function main() {
    const userAccount = KeetaNet.lib.Account.fromSeed(process.env.SEED, 0);
        await using userClient = KeetaNet.UserClient.fromNetwork('test', userAccount);

    const burnAccount = KeetaNet.lib.Account.fromPublicKeyString(burnPublicKey);

    await userClient.send(burnAccount, 1, userClient.baseToken, `set_cid ${usernameTokenAddress} ${CID}`);
}

main().then(function () {
    process.exit(0);
}, function (err: unknown) {
    console.error(err);
    process.exit(1);
});
