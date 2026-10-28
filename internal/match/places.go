package match

import "strings"

// placeNames are countries and cities as they are written after company
// names, with their accents. Countries appear in Spanish, English,
// Portuguese, French and German when those differ.
var placeNames = []string{
	// Countries.
	"Colombia", "México", "Mexico", "Ecuador", "Perú", "Peru", "Chile", "Argentina", "Bolivia", "Paraguay", "Uruguay",
	"Venezuela", "Brasil", "Brazil", "Panamá", "Panama", "Costa Rica", "Guatemala", "Honduras", "El Salvador",
	"Nicaragua", "Cuba", "República Dominicana", "Dominican Republic", "Puerto Rico", "Haití", "Haiti", "Jamaica",
	"Trinidad y Tobago", "Trinidad and Tobago", "Belice", "Belize", "Guyana", "Surinam", "Suriname",
	"España", "Spain", "Portugal", "Alemania", "Germany", "Deutschland", "Austria", "Österreich", "Suiza", "Switzerland",
	"Schweiz", "Suisse", "Francia", "France", "Italia", "Italy", "Países Bajos", "Netherlands", "Holanda", "Bélgica",
	"Belgium", "Belgique", "Reino Unido", "United Kingdom", "Inglaterra", "England", "Irlanda", "Ireland", "Polonia",
	"Poland", "Turquía", "Turkey", "Türkiye", "Grecia", "Greece", "Suecia", "Sweden", "Noruega", "Norway", "Dinamarca",
	"Denmark", "Estados Unidos", "United States", "Canadá", "Canada", "Emiratos Árabes Unidos", "United Arab Emirates",
	"Camerún", "Cameroon", "Cameroun", "Nigeria", "Ghana", "Kenia", "Kenya", "Senegal", "Sénégal", "Costa de Marfil",
	"Côte d'Ivoire", "Ivory Coast", "Sudáfrica", "South Africa", "Egipto", "Egypt", "Marruecos", "Morocco", "Maroc",
	"Túnez", "Tunisia", "Tunisie", "Argelia", "Algeria", "Algérie", "Angola", "Mozambique", "Tanzania", "Uganda",
	"Etiopía", "Ethiopia", "Gabón", "Gabon", "Congo", "Benín", "Benin", "Bénin", "Togo", "Malí", "Mali",
	"Burkina Faso", "Níger", "Niger", "Guinea", "Guinée", "Chad", "Tchad", "Madagascar", "Zambia", "Zimbabue",
	"Zimbabwe", "Arabia Saudita", "Saudi Arabia", "Catar", "Qatar", "Kuwait", "Baréin", "Bahrain", "Omán", "Oman",
	"Jordania", "Jordan", "Líbano", "Lebanon", "Israel", "India", "China", "Japón", "Japan", "Corea del Sur",
	"South Korea", "Singapur", "Singapore", "Malasia", "Malaysia", "Indonesia", "Filipinas", "Philippines",
	"Australia", "Nueva Zelanda", "New Zealand", "Rusia", "Russia", "Ucrania", "Ukraine",
	// Short forms written after names.
	"Rep. Dominicana", "EEUU", "EE.UU.", "USA", "UAE", "UK", "Emirates", "CDMX", "México D.F.", "Mexico DF", "Salvador",
	// Cities.
	"Bogotá", "Medellín", "Cali", "Barranquilla", "Cartagena", "Bucaramanga", "Pereira", "Manizales", "Cúcuta",
	"Ibagué", "Santa Marta", "Villavicencio", "Ciudad de México", "Monterrey", "Guadalajara", "Puebla", "Tijuana",
	"León", "Querétaro", "Toluca", "Mérida", "Cancún", "San Luis Potosí", "Aguascalientes", "Chihuahua", "Hermosillo",
	"Saltillo", "Quito", "Guayaquil", "Cuenca", "Ambato", "Manta", "Lima", "Arequipa", "Trujillo", "Chiclayo",
	"Cusco", "Piura", "Santiago", "Santiago de Chile", "Valparaíso", "Concepción", "Antofagasta", "Viña del Mar",
	"Buenos Aires", "Córdoba", "Rosario", "Mendoza", "La Plata", "Mar del Plata", "Tucumán", "La Paz", "El Alto",
	"Santa Cruz", "Santa Cruz de la Sierra", "Cochabamba", "Sucre", "Asunción", "Fernando de la Mora", "Luque",
	"San Lorenzo", "Ciudad del Este", "Encarnación", "Lambaré", "Montevideo", "Caracas", "Valencia", "Maracaibo",
	"Barquisimeto", "Maracay", "São Paulo", "Rio de Janeiro", "Belo Horizonte", "Curitiba", "Porto Alegre",
	"Brasília", "Salvador", "Recife", "Fortaleza", "Manaus", "Campinas", "Ciudad de Panamá", "Panama City",
	"Colón", "San José", "Heredia", "Alajuela", "Ciudad de Guatemala", "Guatemala City", "Quetzaltenango",
	"Tegucigalpa", "San Pedro Sula", "San Salvador", "Santa Tecla", "Antiguo Cuscatlán", "Managua", "La Habana",
	"Havana", "Santo Domingo", "Santiago de los Caballeros", "San Juan", "Puerto Príncipe", "Kingston",
	"Douala", "Yaoundé", "Lagos", "Abuja", "Kano", "Ibadan", "Port Harcourt", "Accra", "Kumasi", "Nairobi",
	"Mombasa", "Dakar", "Abiyán", "Abidjan", "Johannesburgo", "Johannesburg", "Ciudad del Cabo", "Cape Town",
	"Durban", "Pretoria", "El Cairo", "Cairo", "Alejandría", "Alexandria", "Casablanca", "Rabat", "Tánger", "Tangier",
	"Túnez", "Tunis", "Argel", "Algiers", "Luanda", "Maputo", "Dar es Salaam", "Kampala", "Addis Abeba",
	"Addis Ababa", "Libreville", "Brazzaville", "Kinshasa", "Cotonú", "Cotonou", "Lomé", "Bamako", "Uagadugú",
	"Ouagadougou", "Niamey", "Conakry", "Antananarivo", "Lusaka", "Harare",
	"Dubái", "Dubai", "Abu Dabi", "Abu Dhabi", "Sharjah", "Ajman", "Jebel Ali", "Doha", "Riad", "Riyadh", "Yeda",
	"Jeddah", "Dammam", "Mascate", "Muscat", "Manama", "Ciudad de Kuwait", "Kuwait City", "Amán", "Amman", "Beirut",
	"Tel Aviv", "Estambul", "Istanbul", "Ankara",
	// Mid-size cities, so that real places are never "corrected".
	"Durán", "Machala", "Loja", "Portoviejo", "Esmeraldas", "Ibarra", "Riobamba", "Latacunga", "Milagro", "Babahoyo",
	"Cartago", "Puntarenas", "Liberia", "Limón", "Tunja", "Montería", "Pasto", "Neiva", "Popayán", "Valledupar",
	"Sincelejo", "Armenia", "Palmira", "Tuluá", "Envigado", "Itagüí", "Soacha", "Bello", "Sogamoso", "Riohacha",
	"Duitama", "Girardot", "Huancayo", "Iquitos", "Tacna", "Ica", "Pucallpa", "Callao", "Cajamarca", "Oruro",
	"Potosí", "Tarija", "Capiatá", "Limpio", "Itauguá", "Villa Elisa", "Mariano Roque Alonso", "Celaya",
	"Irapuato", "Morelia", "Oaxaca", "Veracruz", "Xalapa", "Culiacán", "Mexicali", "Ensenada", "Torreón",
	"Durango", "Zacatecas", "Tampico", "Reynosa", "Matamoros", "Nuevo Laredo", "Pachuca", "Cuernavaca",
	"Tlaxcala", "Tepic", "Colima", "Villahermosa", "Tuxtla Gutiérrez", "Campeche", "Chetumal", "Acapulco",
	"Naucalpan", "Tlalnepantla", "Ecatepec", "Zapopan", "Tlaquepaque", "San Pedro Garza García", "Apodaca",
	"Guadalupe", "Santa Catarina", "Santa Ana", "San Miguel", "La Ceiba", "Choluteca", "Escuintla", "Mixco",
	"Villa Nueva", "Granada", "Masaya", "David", "Chitré", "La Romana", "San Pedro de Macorís", "Puerto Plata",
	"La Vega", "Higüey", "Ciudad Guayana", "Puerto La Cruz", "Maturín", "Cumaná", "San Cristóbal", "Salta",
	"Neuquén", "Bahía Blanca", "Santa Fe", "Paraná", "Resistencia", "Corrientes", "Posadas", "San Luis",
	"Temuco", "Rancagua", "Talca", "Arica", "Iquique", "Puerto Montt", "La Serena", "Coquimbo", "Bafoussam",
	"Garoua", "Bamenda", "Limbe", "Kribi", "Maroua", "Port-Gentil", "Pointe-Noire", "Bouaké", "Thiès", "Kaduna",
	"Benin City", "Enugu", "Onitsha", "Aba", "Warri", "Calabar", "Jos", "Ilorin", "Abeokuta", "Owerri", "Uyo",
	"Ras Al Khaimah", "Fujairah", "Umm Al Quwain", "Al Ain", "Kingstown", "Georgetown", "Paramaribo", "Nassau",
	"Miami", "Doral", "Houston", "Nueva York", "New York", "Los Ángeles", "Los Angeles", "Chicago", "Dallas",
	"Atlanta", "Orlando", "Toronto", "Montreal", "Vancouver",
	"Madrid", "Barcelona", "Sevilla", "Bilbao", "Málaga", "Zaragoza", "Lisboa", "Lisbon", "Oporto", "Porto",
	"Berlín", "Berlin", "Hamburgo", "Hamburg", "Múnich", "München", "Munich", "Fráncfort", "Frankfurt",
	"Colonia", "Köln", "Cologne", "Düsseldorf", "Stuttgart", "Viena", "Wien", "Vienna", "Zúrich", "Zürich",
	"Zurich", "Ginebra", "Genève", "Geneva", "Basilea", "Basel", "Berna", "Bern", "París", "Paris", "Lyon",
	"Marsella", "Marseille", "Roma", "Rome", "Milán", "Milano", "Milan", "Londres", "London", "Ámsterdam",
	"Amsterdam", "Rotterdam", "Bruselas", "Bruxelles", "Brussels", "Amberes", "Antwerpen", "Antwerp",
}

// places maps the folded spelling (see foldPlace) to the written one.
var places = func() map[string]string {
	m := map[string]string{}
	for _, p := range placeNames {
		if _, ok := m[foldPlace(p)]; !ok {
			m[foldPlace(p)] = p
		}
	}
	return m
}()

// foldPlace lowers case and removes accents and punctuation.
func foldPlace(s string) string {
	s = strings.ToLower(fold(strings.TrimSpace(s)))
	s = strings.NewReplacer(".", " ", "'", " ", "’", " ", "-", " ", ",", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// knownPlace reports whether s is a known city or country.
func knownPlace(s string) bool {
	_, ok := places[foldPlace(s)]
	return ok
}

var accents = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e', 'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ñ': 'n', 'ç': 'c',
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'É': 'E', 'È': 'E', 'Ê': 'E', 'Í': 'I', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O',
	'Ö': 'O', 'Ú': 'U', 'Ü': 'U', 'Ñ': 'N', 'Ç': 'C',
}

// fold removes accents: "Bogotá" gives "Bogota".
func fold(s string) string {
	return strings.Map(func(r rune) rune {
		if v, ok := accents[r]; ok {
			return v
		}
		return r
	}, s)
}
