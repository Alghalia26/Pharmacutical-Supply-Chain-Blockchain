'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class TraceabilityWorkload extends WorkloadModuleBase {
    async submitTransaction() {
        const batchID = 'TRACE_' + this.workerIndex + '_' + Date.now() + '_' + Math.floor(Math.random() * 1000000000);

        const drugs = ['Panadol', 'Insulin', 'Amoxicillin', 'Ibuprofen', 'Aspirin'];
        const countries = ['Oman', 'Germany', 'India', 'UK', 'UAE'];
        const shippers = ['DHL Oman', 'Aramex Oman', 'Oman Logistics'];
        const distributors = ['Muscat Medical Distributor', 'Oman Pharma Distributor', 'Gulf Medical Supplies'];

        const drugName = drugs[Math.floor(Math.random() * drugs.length)];
        const originCountry = countries[Math.floor(Math.random() * countries.length)];
        const shipperName = shippers[Math.floor(Math.random() * shippers.length)];
        const distributorName = distributors[Math.floor(Math.random() * distributors.length)];

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'createBatch',
            invokerIdentity: 'client0',
            contractArguments: [
                batchID, drugName, originCountry, 'ShippingOrg1', 'At Shipping',
                '2026-01-01', '2027-01-01', '5', '100', shipperName
            ]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'transferBatch',
            invokerIdentity: 'client0',
            contractArguments: [batchID, 'DistributorOrg2', distributorName]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'queryBatch',
            invokerIdentity: 'clientDistributor',
            contractArguments: [batchID],
            readOnly: true
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'transferBatch',
            invokerIdentity: 'clientDistributor',
            contractArguments: [batchID, 'PharmacyOrg3', 'N/A']
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'confirmReceipt',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, '100']
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'getHistoryForBatch',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID],
            readOnly: true
        });
    }
}

function createWorkloadModule() {
    return new TraceabilityWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
