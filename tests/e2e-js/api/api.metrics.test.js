/**
 * Tests for the metrics api endpoint.
 * 
 * Jam-build, a web application practical reference.
 * Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
 * 
 * This file is part of Jam-build.
 * Jam-build is free software: you can redistribute it and/or modify it
 * under the terms of the GNU Affero General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later version.
 * Jam-build is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
 * without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the GNU Affero General Public License for more details.
 * You should have received a copy of the GNU Affero General Public License along with Jam-build.
 * If not, see <https://www.gnu.org/licenses/>.
 * Additional terms under GNU AGPL version 3 section 7:
 * a) The reasonable legal notice of original copyright and author attribution must be preserved
 *    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
 *    in this material, copies, or source code of derived works.
 */
import { test, expect } from '#test/fixtures.js';
import { postData } from './api.js';

test.describe('/api/metrics tests', () => {
  let baseUrl;

  test.beforeAll(() => {
    baseUrl = `${process.env.BASE_URL}/api/metrics`;
  });

  test('POST /api/metrics successfully records a known event', async ({ request }) => {
    await postData(request, baseUrl, {
      event: 'version_conflict_backoff',
      labels: {
        retryCount: 1,
        appVersion: '1.0',
        apiVersion: '1.0.0',
        schemaVersion: '2.0.0'
      }
    }, {
      expectSuccess: true,
      expectResponse: false,
      assertStatus: 204
    });
  });

  test('POST /api/metrics successfully ignores an unknown event but returns 204', async ({ request }) => {
    await postData(request, baseUrl, {
      event: 'unknown_event_type',
      labels: {
        someKey: 'someValue'
      }
    }, {
      expectSuccess: true,
      expectResponse: false,
      assertStatus: 204
    });
  });

  test('POST /api/metrics returns 400 on invalid payload format', async ({ request }) => {
    const response = await request.post(baseUrl, {
      data: 'invalid string payload instead of object'
    });
    
    expect(response.status()).toEqual(400);

    const json = await response.json();
    expect(json.error).toBeDefined();
  });
});
