/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
import type { EnduroStoredPackage } from './EnduroStoredPackage';
import {
    EnduroStoredPackageFromJSON,
    EnduroStoredPackageFromJSONTyped,
    EnduroStoredPackageToJSON,
} from './EnduroStoredPackage';

/**
 * 
 * @export
 * @interface PackageUpdatedEvent
 */
export interface PackageUpdatedEvent {
    /**
     * Identifier of package
     * @type {number}
     * @memberof PackageUpdatedEvent
     */
    id: number;
    /**
     * 
     * @type {EnduroStoredPackage}
     * @memberof PackageUpdatedEvent
     */
    item: EnduroStoredPackage;
}

/**
 * Check if a given object implements the PackageUpdatedEvent interface.
 */
export function instanceOfPackageUpdatedEvent(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "id" in value;
    isInstance = isInstance && "item" in value;

    return isInstance;
}

export function PackageUpdatedEventFromJSON(json: any): PackageUpdatedEvent {
    return PackageUpdatedEventFromJSONTyped(json, false);
}

export function PackageUpdatedEventFromJSONTyped(json: any, ignoreDiscriminator: boolean): PackageUpdatedEvent {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'id': json['id'],
        'item': EnduroStoredPackageFromJSON(json['item']),
    };
}

export function PackageUpdatedEventToJSON(value?: PackageUpdatedEvent | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'id': value.id,
        'item': EnduroStoredPackageToJSON(value.item),
    };
}
