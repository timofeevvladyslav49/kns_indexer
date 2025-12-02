import * as KeetaNet from '@keetanetwork/keetanet-client';

const tokenName = 'KNS';
const username = 'username';

async function main() {
    const userAccount = KeetaNet.lib.Account.fromSeed(process.env.SEED, 0);
    await using userClient = KeetaNet.UserClient.fromNetwork('test', userAccount);

    const {account: token} = await userClient.generateIdentifier(KeetaNet.lib.Account.AccountKeyAlgorithm.TOKEN);
    if (!token.isToken()) {
        throw (new Error('Tokens Should be TOKEN Key Algorithm'));
    }

    const metadata = Buffer.from(JSON.stringify({decimalPlaces: 0}), 'utf-8').toString('base64');

    const builder = userClient.initBuilder();
    builder.setInfo({
        name: tokenName,
        description: username,
        metadata: metadata,
        defaultPermission: new KeetaNet.lib.Permissions(['ACCESS'])
    }, {account: token});

    builder.modifyTokenSupply(1n, {account: token});

    await builder.computeBlocks();

    builder.send(userAccount, 1n, token, undefined, {account: token});

    await builder.computeBlocks();

    await builder.publish();
}

main().then(function () {
    process.exit(0);
}, function (err: unknown) {
    console.error(err);
    process.exit(1);
});
