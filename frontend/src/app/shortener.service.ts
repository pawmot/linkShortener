import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ShortenerService {

  constructor(private http: HttpClient) {
  }

  shorten(link: string): Promise<string> {
    return this.http.post<ShortenerResponse>(
      environment.shortenerServiceUrl,
      {
        url: link
      })
      .toPromise()
      .then(res => res.shortenedLink);
  }
}

interface ShortenerResponse {
  shortenedLink: string;
}
