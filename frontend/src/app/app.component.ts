import { Component } from '@angular/core';
import {ShortenerService} from './shortener.service';
import {Clipboard} from '@angular/cdk/clipboard';
import {MatSnackBar} from '@angular/material/snack-bar';

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
    private shortenerSvc: ShortenerService,
    private clipboard: Clipboard,
    private snackBar: MatSnackBar) {
  }

  async shorten(): Promise<void> {
     this.shortenedLink = await this.shortenerSvc.shorten(this.link);
     this.shortened = true;
  }

  copy(): void {
    this.clipboard.copy(this.shortenedLink);
    this.snackBar.open('Copied!', 'Dismiss', {
      duration: 3000
    });
  }
}
