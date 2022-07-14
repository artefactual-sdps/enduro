/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
import {
    EnduroPackagePreservationActionResponseBodySimple,
    EnduroPackagePreservationActionResponseBodySimpleFromJSON,
    EnduroPackagePreservationActionResponseBodySimpleFromJSONTyped,
    EnduroPackagePreservationActionResponseBodySimpleToJSON,
} from './EnduroPackagePreservationActionResponseBodySimple';

/**
 * EnduroPreservation-Action-Created-EventResponseBody result type (default view)
 * @export
 * @interface EnduroPreservationActionCreatedEventResponseBody
 */
export interface EnduroPreservationActionCreatedEventResponseBody {
    /**
     * Identifier of preservation action
     * @type {number}
     * @memberof EnduroPreservationActionCreatedEventResponseBody
     */
    id: number;
    /**
     * 
     * @type {EnduroPackagePreservationActionResponseBodySimple}
     * @memberof EnduroPreservationActionCreatedEventResponseBody
     */
    item: EnduroPackagePreservationActionResponseBodySimple;
}

export function EnduroPreservationActionCreatedEventResponseBodyFromJSON(json: any): EnduroPreservationActionCreatedEventResponseBody {
    return EnduroPreservationActionCreatedEventResponseBodyFromJSONTyped(json, false);
}

export function EnduroPreservationActionCreatedEventResponseBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): EnduroPreservationActionCreatedEventResponseBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'id': json['id'],
        'item': EnduroPackagePreservationActionResponseBodySimpleFromJSON(json['item']),
    };
}

export function EnduroPreservationActionCreatedEventResponseBodyToJSON(value?: EnduroPreservationActionCreatedEventResponseBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'id': value.id,
        'item': EnduroPackagePreservationActionResponseBodySimpleToJSON(value.item),
    };
}

