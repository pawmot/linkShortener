import { Component } from '@angular/core';
import {ShortenerService} from './shortener.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  shortened = false;
  title = 'linkShortener';
  link = '';
  shortenedLink = '';

  constructor(
    private shortenerSvc: ShortenerService) {
  }

  async shorten(): Promise<void> {
     this.shortenedLink = await this.shortenerSvc.shorten(this.link);
     this.shortened = true;
  }
}
