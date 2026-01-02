import http from 'k6/http';
import { check } from 'k6';
// const uuid = generateUUID();

export const options = {
    executor: 'constant-vus',
    vus: 1000,            // 100 users
    duration: '10s',     // run for 10 seconds
    startTime: '0s',     // all start at once
};
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

console.log(generateUUID());

export default function () {
    const url = 'http://core-api.tm-vault-sit.svc.cluster.local:8080/v1/posting-instruction-batches';

    const payload = JSON.stringify({
        request_id: generateUUID(),
        posting_instruction_batch: {
            client_id: "AsyncCreatePostingInstructionBatch",
            client_batch_id: generateUUID(),
            posting_instructions: [
                {
                    client_transaction_id: generateUUID(),
                    inbound_hard_settlement: {
                        amount: "10",
                        denomination: "THB",
                        target_account: {
                            account_id: "8fe1b313-2d95-4a13-b3bb-8f2851c8493d"
                        },
                        internal_account_id: "1"
                    },
                    instruction_details: {
                        transaction_type: "fund_transfer",
                        transaction_code: "SSAMAFN",
                        transaction_class: "D"
                    }
                },
                {
                    client_transaction_id: generateUUID(),
                    inbound_hard_settlement: {
                        amount: "10",
                        denomination: "THB",
                        target_account: {
                            account_id: "bcdfa76d-3695-4db2-9a46-c3b26865c960"
                        },
                        internal_account_id: "1"
                    },
                    instruction_details: {
                        transaction_type: "fund_transfer",
                        transaction_code: "SSAMAFN",
                        transaction_class: "D"
                    }
                }
            ],
            batch_details: {}
        }
    });

    const headers = {
        'X-Auth-Token': 'A0001200834200745148371!PNo1WzYGqW2DtBZNyaBxTrXHW8CW6NejIccqTDhfRXtWk6q5LnFPhvaieUPq/3AFgnkN7OU1Uv7Aood8k/JznmAq0iI=',
        'Content-Type': 'application/json',
    };

    const res = http.post(url, payload, { headers });

    check(res, {
        'is status 200 or 201': (r) => r.status === 200 || r.status === 201,
        'response has client_batch_id': (r) => r.body.includes('client_batch_id'),
    });
}
