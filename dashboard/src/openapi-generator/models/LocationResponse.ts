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
/**
 * A Location describes a location retrieved by the storage service. (default view)
 * @export
 * @interface LocationResponse
 */
export interface LocationResponse {
    /**
     * Creation datetime
     * @type {Date}
     * @memberof LocationResponse
     */
    createdAt: Date;
    /**
     * Description of the location
     * @type {string}
     * @memberof LocationResponse
     */
    description?: string;
    /**
     * Name of location
     * @type {string}
     * @memberof LocationResponse
     */
    name: string;
    /**
     * Purpose of the location
     * @type {string}
     * @memberof LocationResponse
     */
    purpose: LocationResponsePurposeEnum;
    /**
     * Data source of the location
     * @type {string}
     * @memberof LocationResponse
     */
    source: LocationResponseSourceEnum;
    /**
     * 
     * @type {string}
     * @memberof LocationResponse
     */
    uuid: string;
}


/**
 * @export
 */
export const LocationResponsePurposeEnum = {
    Unspecified: 'unspecified',
    AipStore: 'aip_store'
} as const;
export type LocationResponsePurposeEnum = typeof LocationResponsePurposeEnum[keyof typeof LocationResponsePurposeEnum];

/**
 * @export
 */
export const LocationResponseSourceEnum = {
    Unspecified: 'unspecified',
    Minio: 'minio'
} as const;
export type LocationResponseSourceEnum = typeof LocationResponseSourceEnum[keyof typeof LocationResponseSourceEnum];


/**
 * Check if a given object implements the LocationResponse interface.
 */
export function instanceOfLocationResponse(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "createdAt" in value;
    isInstance = isInstance && "name" in value;
    isInstance = isInstance && "purpose" in value;
    isInstance = isInstance && "source" in value;
    isInstance = isInstance && "uuid" in value;

    return isInstance;
}

export function LocationResponseFromJSON(json: any): LocationResponse {
    return LocationResponseFromJSONTyped(json, false);
}

export function LocationResponseFromJSONTyped(json: any, ignoreDiscriminator: boolean): LocationResponse {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'createdAt': (new Date(json['created_at'])),
        'description': !exists(json, 'description') ? undefined : json['description'],
        'name': json['name'],
        'purpose': json['purpose'],
        'source': json['source'],
        'uuid': json['uuid'],
    };
}

export function LocationResponseToJSON(value?: LocationResponse | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'created_at': (value.createdAt.toISOString()),
        'description': value.description,
        'name': value.name,
        'purpose': value.purpose,
        'source': value.source,
        'uuid': value.uuid,
    };
}

