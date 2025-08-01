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
 * @interface StoragePingEvent
 */
export interface StoragePingEvent {
    /**
     * 
     * @type {string}
     * @memberof StoragePingEvent
     */
    message?: string;
}

/**
 * Check if a given object implements the StoragePingEvent interface.
 */
export function instanceOfStoragePingEvent(value: object): boolean {
    let isInstance = true;

    return isInstance;
}

export function StoragePingEventFromJSON(json: any): StoragePingEvent {
    return StoragePingEventFromJSONTyped(json, false);
}

export function StoragePingEventFromJSONTyped(json: any, ignoreDiscriminator: boolean): StoragePingEvent {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'message': !exists(json, 'message') ? undefined : json['message'],
    };
}

export function StoragePingEventToJSON(value?: StoragePingEvent | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'message': value.message,
    };
}

