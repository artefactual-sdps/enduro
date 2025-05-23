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
 * SIP not found
 * @export
 * @interface SIPNotFound
 */
export interface SIPNotFound {
    /**
     * Message of error
     * @type {string}
     * @memberof SIPNotFound
     */
    message: string;
    /**
     * Identifier of missing SIP
     * @type {string}
     * @memberof SIPNotFound
     */
    uuid: string;
}

/**
 * Check if a given object implements the SIPNotFound interface.
 */
export function instanceOfSIPNotFound(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "message" in value;
    isInstance = isInstance && "uuid" in value;

    return isInstance;
}

export function SIPNotFoundFromJSON(json: any): SIPNotFound {
    return SIPNotFoundFromJSONTyped(json, false);
}

export function SIPNotFoundFromJSONTyped(json: any, ignoreDiscriminator: boolean): SIPNotFound {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'message': json['message'],
        'uuid': json['uuid'],
    };
}

export function SIPNotFoundToJSON(value?: SIPNotFound | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'message': value.message,
        'uuid': value.uuid,
    };
}

