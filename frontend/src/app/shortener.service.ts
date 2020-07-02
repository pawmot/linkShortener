import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ShortenerService {

  constructor(private http: HttpClient) {
  }

  shorten(link: string): Promise<string> {
    return this.http.post<ShortenerResponse>(
      'https://1cpmx32uyj.execute-api.eu-west-1.amazonaws.com/registerLink', {
      url: link
    }).toPromise()
      .then(res => res.shortenedLink);
  }
}

interface ShortenerResponse {
  shortenedLink: string;
}
