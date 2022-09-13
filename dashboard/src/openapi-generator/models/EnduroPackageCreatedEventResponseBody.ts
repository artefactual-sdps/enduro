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
import type { EnduroStoredPackageResponseBody } from './EnduroStoredPackageResponseBody';
import {
    EnduroStoredPackageResponseBodyFromJSON,
    EnduroStoredPackageResponseBodyFromJSONTyped,
    EnduroStoredPackageResponseBodyToJSON,
} from './EnduroStoredPackageResponseBody';

/**
 * EnduroPackage-Created-EventResponseBody result type (default view)
 * @export
 * @interface EnduroPackageCreatedEventResponseBody
 */
export interface EnduroPackageCreatedEventResponseBody {
    /**
     * Identifier of package
     * @type {number}
     * @memberof EnduroPackageCreatedEventResponseBody
     */
    id: number;
    /**
     * 
     * @type {EnduroStoredPackageResponseBody}
     * @memberof EnduroPackageCreatedEventResponseBody
     */
    item: EnduroStoredPackageResponseBody;
}

/**
 * Check if a given object implements the EnduroPackageCreatedEventResponseBody interface.
 */
export function instanceOfEnduroPackageCreatedEventResponseBody(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "id" in value;
    isInstance = isInstance && "item" in value;

    return isInstance;
}

export function EnduroPackageCreatedEventResponseBodyFromJSON(json: any): EnduroPackageCreatedEventResponseBody {
    return EnduroPackageCreatedEventResponseBodyFromJSONTyped(json, false);
}

export function EnduroPackageCreatedEventResponseBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): EnduroPackageCreatedEventResponseBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'id': json['id'],
        'item': EnduroStoredPackageResponseBodyFromJSON(json['item']),
    };
}

export function EnduroPackageCreatedEventResponseBodyToJSON(value?: EnduroPackageCreatedEventResponseBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'id': value.id,
        'item': EnduroStoredPackageResponseBodyToJSON(value.item),
    };
}

