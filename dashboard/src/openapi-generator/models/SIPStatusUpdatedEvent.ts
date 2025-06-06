/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 0.0.1
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface SIPStatusUpdatedEvent
 */
export interface SIPStatusUpdatedEvent {
    /**
     * 
     * @type {string}
     * @memberof SIPStatusUpdatedEvent
     */
    status: SIPStatusUpdatedEventStatusEnum;
    /**
     * Identifier of SIP
     * @type {string}
     * @memberof SIPStatusUpdatedEvent
     */
    uuid: string;
}


/**
 * @export
 */
export const SIPStatusUpdatedEventStatusEnum = {
    Error: 'error',
    Failed: 'failed',
    Queued: 'queued',
    Processing: 'processing',
    Pending: 'pending',
    Ingested: 'ingested'
} as const;
export type SIPStatusUpdatedEventStatusEnum = typeof SIPStatusUpdatedEventStatusEnum[keyof typeof SIPStatusUpdatedEventStatusEnum];


/**
 * Check if a given object implements the SIPStatusUpdatedEvent interface.
 */
export function instanceOfSIPStatusUpdatedEvent(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "status" in value;
    isInstance = isInstance && "uuid" in value;

    return isInstance;
}

export function SIPStatusUpdatedEventFromJSON(json: any): SIPStatusUpdatedEvent {
    return SIPStatusUpdatedEventFromJSONTyped(json, false);
}

export function SIPStatusUpdatedEventFromJSONTyped(json: any, ignoreDiscriminator: boolean): SIPStatusUpdatedEvent {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'status': json['status'],
        'uuid': json['uuid'],
    };
}

export function SIPStatusUpdatedEventToJSON(value?: SIPStatusUpdatedEvent | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'status': value.status,
        'uuid': value.uuid,
    };
}

