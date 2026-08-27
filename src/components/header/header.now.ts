// header.now.ts
import { Component } from '@nowts/core';

@Component({
  selector: 'app-header',
  templateUrl: './header.now.html',
  styleUrl: './header.now.css'
})
export class HeaderComponent {
  title = 'header component';

  constructor() {}
}
